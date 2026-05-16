package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-oauth-login/env"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

/*
 * @Desc: etcd服务
 * @author: 福狼
 * @version: v1.0.0
 */

// 全局etcd客户端
var (
	etcdClient *clientv3.Client
	once       sync.Once
	mu         sync.Mutex
)

// 初始化etcd客户端
func initEtcdClient() error {
	mu.Lock()
	defer mu.Unlock()

	var err error
	once.Do(func() {
		// 读取etcd地址
		etcdHost := env.GetServerConfig().Etcd.EtcdHost
		etcdPort := env.GetServerConfig().Etcd.EtcdPort
		addr := fmt.Sprintf("%s:%s", etcdHost, etcdPort)

		// 创建etcd客户端
		etcdClient, err = clientv3.New(clientv3.Config{
			Endpoints:            []string{addr},   // 支持集群地址 ["etcd1:2379", "etcd2:2379"]
			DialTimeout:          5 * time.Second,  // 连接超时
			DialKeepAliveTime:    30 * time.Second, // 保活时间
			DialKeepAliveTimeout: 10 * time.Second, // 保活超时
		})
	})

	if err != nil {
		return fmt.Errorf("初始化etcd客户端失败: %w", err)
	}
	if etcdClient == nil {
		return fmt.Errorf("etcd客户端未初始化")
	}
	return nil
}

// ========================== 服务发现：实现gRPC resolver接口 ==========================

// etcdResolver 实现gRPC的resolver.Resolver接口
type etcdResolver struct {
	client *clientv3.Client
	target resolver.Target
	cc     resolver.ClientConn
	ctx    context.Context
	cancel context.CancelFunc
}

// etcdBuilder 实现gRPC的resolver.Builder接口
type etcdBuilder struct {
	cli *clientv3.Client
}

// InitEtcdOnStart 初始化etcd解析器
func InitEtcdOnStart() error {
	// 确保etcd客户端已初始化
	if err := initEtcdClient(); err != nil {
		return fmt.Errorf("初始化etcd客户端失败: %w", err)
	}

	// 注册解析器
	resolver.Register(&etcdBuilder{cli: etcdClient})
	log.Println("etcd解析器初始化成功")
	return nil
}

// Build 创建解析器实例
func (b *etcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &etcdResolver{
		client: b.cli,
		target: target,
		cc:     cc,
		ctx:    ctx,
		cancel: cancel,
	}

	// 异步监听etcd
	go r.watch()
	return r, nil
}

// Scheme 返回解析器的scheme 通过"etcd:///"指定
func (b *etcdBuilder) Scheme() string {
	return "etcd"
}

// watch 监听etcd中服务地址的变化
func (r *etcdResolver) watch() {
	prefix := "/services/" + r.target.Endpoint() + "/"
	log.Printf("开始监听etcd服务路径: %s", prefix)

	// 循环监听
	for {
		select {
		case <-r.ctx.Done():
			log.Printf("解析器已关闭，停止监听: %s", prefix)
			return
		default:
			// 失败重试
			if err := r.watchOnce(prefix); err != nil {
				log.Printf("etcd监听异常，将在3秒后重试: %v", err)
				time.Sleep(3 * time.Second)
			}
		}
	}
}

// watchOnce 单次监听逻辑
func (r *etcdResolver) watchOnce(prefix string) error {
	// 初始查询etcd中的服务地址
	resp, err := r.client.Get(r.ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("初始查询服务地址失败: %w", err)
	}

	// 解析初始地址
	addrs := r.parseAddresses(resp.Kvs, prefix)
	if err := r.updateClientConn(addrs); err != nil {
		return fmt.Errorf("初始更新服务地址失败: %w", err)
	}

	// 监听后续地址变化
	watcher := r.client.Watch(r.ctx, prefix, clientv3.WithPrefix())
	for {
		select {
		case <-r.ctx.Done():
			return nil
		case resp, ok := <-watcher:
			if !ok {
				return fmt.Errorf("watch通道已关闭")
			}
			if resp.Err() != nil {
				return resp.Err()
			}
			// 处理事件
			addrs = r.handleEvents(resp.Events, addrs, prefix)
			if err := r.updateClientConn(addrs); err != nil {
				log.Printf("更新服务地址失败: %v", err) // 非致命错误，继续监听
			}
		}
	}
}

// parseAddresses 从etcd的Kv列表中解析服务地址
func (r *etcdResolver) parseAddresses(kvs []*mvccpb.KeyValue, prefix string) []resolver.Address {
	var addrs []resolver.Address
	for _, kv := range kvs {
		addr := strings.TrimPrefix(string(kv.Key), prefix)
		if addr != "" {
			addrs = append(addrs, resolver.Address{Addr: addr})
			log.Printf("发现服务地址: %s", addr)
		}
	}
	if len(addrs) == 0 {
		log.Printf("未发现任何服务地址，路径: %s", prefix)
	}
	return addrs
}

// handleEvents 处理etcd的watch事件
func (r *etcdResolver) handleEvents(events []*clientv3.Event, currentAddrs []resolver.Address, prefix string) []resolver.Address {
	addrs := make([]resolver.Address, len(currentAddrs))
	copy(addrs, currentAddrs)

	for _, ev := range events {
		addr := strings.TrimPrefix(string(ev.Kv.Key), prefix)
		if addr == "" {
			continue
		}

		if ev.Type == clientv3.EventTypePut {
			// 新增地址
			exists := false
			for _, a := range addrs {
				if a.Addr == addr {
					exists = true
					break
				}
			}
			if !exists {
				addrs = append(addrs, resolver.Address{Addr: addr})
				log.Printf("新增服务地址: %s", addr)
			}
		} else {
			// 移除地址
			for i, a := range addrs {
				if a.Addr == addr {
					addrs = append(addrs[:i], addrs[i+1:]...)
					log.Printf("移除服务地址: %s", addr)
					break
				}
			}
		}
	}
	return addrs
}

// updateClientConn 更新gRPC客户端的服务地址
func (r *etcdResolver) updateClientConn(addrs []resolver.Address) error {
	return r.cc.UpdateState(resolver.State{
		Addresses: addrs,
	})
}

// ResolveNow 触发立即解析
func (r *etcdResolver) ResolveNow(opts resolver.ResolveNowOptions) {
	log.Printf("触发立即解析服务: %s", r.target.Endpoint())
}

// Close 关闭解析器
func (r *etcdResolver) Close() {
	r.cancel()
	log.Printf("关闭服务解析器: %s", r.target.Endpoint())
}

// ========================== 服务注册与注销 ==========================

// RegisterToEtcd 注册服务到etcd
func RegisterToEtcd(serviceName, addr string) error {
	if err := initEtcdClient(); err != nil {
		return fmt.Errorf("注册失败：%w", err)
	}

	key := fmt.Sprintf("/services/%s/%s", serviceName, addr)
	log.Printf("开始注册服务到etcd: %s", key)

	// 创建租约 10秒过期
	lease, err := etcdClient.Grant(context.Background(), 10)
	if err != nil {
		return fmt.Errorf("创建租约失败: %w", err)
	}

	// 注册服务
	_, err = etcdClient.Put(context.Background(), key, "", clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("写入etcd失败: %w", err)
	}

	// 启动续约监控
	go keepAliveLoop(key, lease.ID)
	return nil
}

// keepAliveLoop 租约续约
func keepAliveLoop(key string, leaseID clientv3.LeaseID) {
	for {
		// 启动续约
		keepAliveChan, err := etcdClient.KeepAlive(context.Background(), leaseID)
		if err != nil {
			log.Printf("续约失败，将重新注册服务: %v", err)
			time.Sleep(2 * time.Second)

			// 重新创建租约并注册
			newLease, err := etcdClient.Grant(context.Background(), 10)
			if err != nil {
				log.Printf("重新创建租约失败: %v", err)
				continue
			}
			leaseID = newLease.ID
			if _, err := etcdClient.Put(context.Background(), key, "", clientv3.WithLease(leaseID)); err != nil {
				log.Printf("重新注册服务失败: %v", err)
				continue
			}
			log.Printf("服务已重新注册: %s", key)
			continue
		}

		// 监听续约响应
		log.Printf("服务续约已启动: %s", key)
		for ka := range keepAliveChan {
			if ka == nil {
				log.Printf("续约通道关闭，准备重新注册: %s", key)
				break
			}
			// 定期打印续约日志
			// log.Printf("etcd续约成功: %+v", ka)
		}
	}
}

// UnregisterFromEtcd 从etcd注销服务
func UnregisterFromEtcd(serviceName, addr string) error {
	if err := initEtcdClient(); err != nil {
		return fmt.Errorf("注销失败：%w", err)
	}

	key := fmt.Sprintf("/services/%s/%s", serviceName, addr)
	_, err := etcdClient.Delete(context.Background(), key)
	if err != nil {
		return fmt.Errorf("删除etcd键失败: %w", err)
	}
	log.Printf("服务已从etcd注销: %s", key)
	return nil
}

// CloseEtcdClient 关闭etcd客户端
func CloseEtcdClient() error {
	if etcdClient != nil {
		return etcdClient.Close()
	}
	return nil
}
