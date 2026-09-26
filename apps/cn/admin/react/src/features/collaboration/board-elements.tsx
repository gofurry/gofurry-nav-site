import { memo, useContext, type CSSProperties } from 'react'
import { BaseEdge, EdgeLabelRenderer, Handle, NodeResizer, NodeToolbar, Position, getBezierPath, getSmoothStepPath, type EdgeProps, type NodeProps } from '@xyflow/react'
import { ArrowSquareOut, PencilSimple, Stack, Trash } from '@phosphor-icons/react'
import { Button } from '../../components/ui/button'
import { boardKinds, colorValues, type CanvasEdge, type CanvasNode } from './board-types'
import { BoardActions } from './board-context'
import { statusLabels, type Status } from './types'

export const CanvasNodeView = memo(function CanvasNodeView({ id, data, selected }: NodeProps<CanvasNode>) {
  const actions = useContext(BoardActions)!
  const node = data.record, ref = data.reference, kind = node.kind
  const isCard = kind === 'note' || kind === 'card'
  const title = ref?.missing ? '引用内容已删除' : ref?.title || node.title
  return <>
    <NodeResizer isVisible={Boolean(selected && actions.writable)} minWidth={isCard ? 200 : kind === 'text' ? 96 : 48} minHeight={isCard ? 160 : kind === 'text' ? 64 : 40} maxWidth={1600} maxHeight={1600} lineStyle={{ borderColor: 'transparent' }} handleStyle={{ backgroundColor: 'var(--primary)', border: 0, borderRadius: 2 }} onResizeStart={() => actions.beginResize(id)} onResizeEnd={(_event, params) => actions.resize(id, params)} />
    <NodeToolbar isVisible={selected} position={Position.Top} className="canvas-node-tools nodrag nopan">
      {actions.writable && <><Button variant="ghost" size="icon" aria-label="编辑元素" title="编辑" onClick={() => actions.editNode(node)}><PencilSimple className="size-4" /></Button><Button variant="ghost" size="icon" aria-label="置顶元素" title="置顶" onClick={() => actions.raise(id)}><Stack className="size-4" /></Button><Button variant="ghost" size="icon" aria-label="删除元素" title="删除" onClick={() => actions.deleteNode(node)}><Trash className="size-4" /></Button></>}
      {ref && !ref.missing && <Button variant="ghost" size="icon" aria-label="打开引用内容" title="打开引用内容" onClick={() => actions.openReference(ref)}><ArrowSquareOut className="size-4" /></Button>}
    </NodeToolbar>
    <article aria-label={`${boardKinds[kind]} ${node.id}`} className={`canvas-element canvas-${kind}${selected ? ' is-selected' : ''}`} style={{ '--node-color': colorValues[node.color] } as CSSProperties}>
      {isCard && <><div className="canvas-card-type"><span />{kind === 'note' ? '便签' : ref ? ref.kind === 'idea' ? '内容想法' : ref.kind === 'game' ? '正式游戏' : '正式网站' : '卡片'}{ref?.status && <span className="ml-auto">{statusLabels[ref.status as Status] || ref.status}</span>}</div><h3 className="line-clamp-2" title={title}>{title || '未命名'}</h3><p className="canvas-card-body">{node.body || (ref ? '双击编辑卡片备注' : '双击编辑内容')}</p>{ref?.missing && <p className="text-xs text-warning">卡片仍保留，可重新选择引用。</p>}</>}
      {kind === 'text' && <><h3>{node.title}</h3><p>{node.body}</p></>}
      {(kind === 'rectangle' || kind === 'ellipse') && <span className="canvas-shape-label">{node.title}</span>}
      {kind === 'arrow' && <><svg viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true" className="h-full w-full"><g transform={`rotate(${node.rotation} 50 50)`}><path d="M10 50 H87 M65 28 L87 50 L65 72" fill="none" stroke="var(--node-color)" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" vectorEffect="non-scaling-stroke" /></g></svg>{node.title && <span className="canvas-arrow-label">{node.title}</span>}</>}
      {isCard && [Position.Top, Position.Right, Position.Bottom, Position.Left].map((position) => <Handle key={position} id={position} type="source" position={position} isConnectable={actions.writable} aria-label={`${boardKinds[kind]} ${node.id} ${position} 连接点`} />)}
    </article>
  </>
})
export function CanvasEdgeView(props: EdgeProps<CanvasEdge>) {
  const actions = useContext(BoardActions)!, edge = props.data!.record
  const path = edge.routing === 'step' ? getSmoothStepPath(props) : getBezierPath(props)
  return <><BaseEdge id={props.id} path={path[0]} markerEnd={props.markerEnd} interactionWidth={24} style={{ stroke: colorValues[edge.color], strokeWidth: props.selected ? 3 : 2 }} /><EdgeLabelRenderer>{(edge.label || props.selected) && <div className="canvas-edge-label nodrag nopan" style={{ transform: `translate(-50%, -50%) translate(${path[1]}px,${path[2]}px)` }}>{edge.label && <span className="max-w-52 truncate" title={edge.label}>{edge.label}</span>}{props.selected && actions.writable && <><Button size="icon" variant="ghost" aria-label="编辑连线" onClick={() => actions.editEdge(edge)}><PencilSimple className="size-4" /></Button><Button size="icon" variant="ghost" aria-label="删除连线" onClick={() => actions.deleteEdge(edge)}><Trash className="size-4" /></Button></>}</div>}</EdgeLabelRenderer></>
}
