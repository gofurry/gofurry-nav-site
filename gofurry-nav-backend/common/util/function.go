package util

/*
 * @Desc: 工具类
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/common"
	cm "github.com/gofurry/gofurry-nav-backend/common/models"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

var clusterId, _ = snowflake.NewNode(int64(env.GetServerConfig().ClusterId))

var letters = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d",
	"e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v",
	"w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N",
	"O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
}

// 雪花算法生成新 ID
func GenerateId() int64 {
	id := clusterId.Generate()
	return id.Int64()
}

// MD5 加密
func CreateMD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func GetClientIP(c fiber.Ctx) string {
	remoteIP := strings.TrimSpace(c.IP())
	if !isTrustedProxyIP(remoteIP, env.GetServerConfig().Server.TrustedProxyCIDRs) {
		return remoteIP
	}

	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		if value := strings.TrimSpace(c.Get(header)); value != "" {
			if header == "X-Forwarded-For" {
				if idx := strings.Index(value, ","); idx >= 0 {
					value = strings.TrimSpace(value[:idx])
				}
			}
			if net.ParseIP(value) != nil {
				return value
			}
		}
	}

	return remoteIP
}

func isTrustedProxyIP(remoteIP string, trustedCIDRs string) bool {
	ip := net.ParseIP(strings.TrimSpace(remoteIP))
	if ip == nil {
		return false
	}
	for _, item := range strings.Split(trustedCIDRs, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if item == "loopback" && ip.IsLoopback() {
			return true
		}
		if _, network, err := net.ParseCIDR(item); err == nil && network.Contains(ip) {
			return true
		}
		if single := net.ParseIP(item); single != nil && single.Equal(ip) {
			return true
		}
	}
	return false
}

// 判断是否为数字
func IsNumber(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}

// 整形返回时间戳
func GetDigitNow() int {
	now, _ := strconv.Atoi(time.Now().Format(common.TIME_FORMAT_DIGIT))
	return now
}

// 字符串返回时间戳
func GetStrNow() string {
	now := time.Now().Format(common.TIME_FORMAT_DIGIT)
	return now
}

// 格式化时间
func GetDateFormatStr(format string, date time.Time) string {
	dataString := date.Format(format)
	return dataString
}

// 字符串转 int
func String2Int(numString string) (int, error) {
	if strings.TrimSpace(numString) == "" {
		return 0, errors.New("字符串不能为0")
	}
	id, err := strconv.Atoi(numString)
	return id, err
}

// int64 转字符串
func Int642String(i64 int64) string { return strconv.FormatInt(i64, 10) }

// int 转字符串
func Int2String(i int) string { return fmt.Sprintf("%d", i) }

// 字符串转 int64
func String2Int64(numString string) (int64, error) {
	if strings.TrimSpace(numString) == "" {
		return 0, errors.New("参数不能为空")
	}
	id, parseErr := strconv.ParseInt(strings.TrimSpace(numString), 10, 64)
	return id, parseErr
}

// 字符串转float64
func String2Float64(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

// 字符串Unicode转float64
func StringUnicode2Float64(str string) []float64 {
	var floatValues []float64
	// 遍历字符串的每个字符
	for _, char := range str {
		// 获取字符的 Unicode 码值
		codePoint := float64(char)
		// 将码值加入 float64 数组
		floatValues = append(floatValues, codePoint)
	}
	return floatValues
}

// float64 转字符串
func Float642String(f64 float64) string { return fmt.Sprintf("%.0f", f64) }

// 大数
func Decimal(num float64) float64 {
	num, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", num), 10)
	return num
}

// 去除末尾字符并转数字
func ExtractSuffix2Int(delayStr string, suffix string) int {
	numStr := strings.TrimSuffix(strings.ToLower(delayStr), suffix)
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return num
}

// 转换时间戳为时间
func UnixToTime(num int64) string {
	return time.Unix(num, 0).Format(common.TIME_FORMAT_DATE)
}

// Is T In List
func In[T comparable](target T, aimList []T) bool {
	for _, item := range aimList {
		if item == target {
			return true
		}
	}
	return false
}

// 生成随机验证码
func GenerateRandomCode(length int) string {
	//rand.Seed(time.Now().UnixNano())
	randCode := ""
	for i := 1; i <= length; i++ {
		randCode += letters[rand.Intn(len(letters))]
	}
	return randCode
}

// JWT 密钥
func Secret() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret()), nil
	}
}

func jwtSecret() string {
	if value := strings.TrimSpace(os.Getenv("GF_NAV_BACKEND_JWT_SECRET")); value != "" {
		return value
	}
	if value := strings.TrimSpace(env.GetServerConfig().Security.JWTSecret); value != "" {
		return value
	}
	return common.TOKEN_SECRET
}

// 解密JWT Token
func ParseToken(authorization string) (*cm.GFClaims, error) {
	token, err := jwt.ParseWithClaims(authorization, &cm.GFClaims{}, Secret())
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, errors.New("token is nil")
	}
	if claims, ok := token.Claims.(*cm.GFClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

/*
iss (issuer): 签发人
exp (expiration time): 过期时间
sub (subject): 主题
aud (audience): 受众
nbf (Not Before): 生效时间
iat (Issued At): 签发时间
jti (JWT ID): 编号
*/
func NewToken(userId string, userName string) (string, error) {
	claims := cm.GFClaims{
		UserId:   userId,
		UserName: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(common.JWT_RELET_NUM * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret()))
}

// 判断是否 IP
func IsIP(ip string) (net.IP, error) {
	addr := net.ParseIP(ip)
	if addr == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}
	return addr, nil
}

// 转换IPv6为IPv4
func ParseIPv6ToIPv4(ip string) (string, error) {
	addr, err := IsIP(ip)
	if err != nil {
		return "", err
	}

	// 如果是IPv4-mapped IPv6地址，则转换为IPv4格式的字符串
	if v4 := addr.To4(); v4 != nil {
		return v4.String(), nil
	}

	// 如果是真正的IPv6地址，检查是否兼容IPv4
	if !IsIPv6LinkLocal(addr) {
		return "", fmt.Errorf("not an IPv6 link-local address: %s", ip)
	}

	// 假设IPv6地址的前缀是fe80::，并且它的后128位是IPv4地址
	if !addr.IsLinkLocalUnicast() {
		return "", fmt.Errorf("not a link-local unicast address: %s", ip)
	}

	return fmt.Sprintf("%d.%d.%d.%d", addr[12], addr[13], addr[14], addr[15]), nil
}

func IsIPv6LinkLocal(ip net.IP) bool {
	return ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast()
}

// 数组去重
func MergeAndDeduplicate(arr1, arr2 []int64) []int64 {
	// 使用 map 去重
	uniqueMap := make(map[int64]struct{})

	// 将第一个数组的元素加入 map
	for _, num := range arr1 {
		uniqueMap[num] = struct{}{}
	}

	// 将第二个数组的元素加入 map
	for _, num := range arr2 {
		uniqueMap[num] = struct{}{}
	}

	// 将 map 的键转为去重后的数组
	uniqueArr := make([]int64, 0, len(uniqueMap))
	for num := range uniqueMap {
		uniqueArr = append(uniqueArr, num)
	}

	return uniqueArr
}

func DecryptPassword(encryptedPassword string, privateKeyPath string) (string, error) {
	// 读取私钥文件
	privKeyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", err
	}
	// 解码 PEM 格式的私钥
	block, _ := pem.Decode(privKeyData)
	if block == nil {
		return "", errors.New("failed to parse PEM block containing the private key")
	}
	// 解析私钥
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	// 将 privKey 转换为 *rsa.PrivateKey
	rsaPrivKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("private key is not of type *rsa.PrivateKey")
	}

	// 解码 Base64 密文
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", errors.New("Base64 解码失败")
	}

	// 解密密码
	decrypted, err := rsa.DecryptPKCS1v15(nil, rsaPrivKey, encryptedData)
	if err != nil {
		return "", err
	}
	// 返回解密后的密码
	return string(decrypted), nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err) // 不存在返回 false，存在返回 true
}
