import type { Node, Edge } from '@xyflow/react'

export type BoardKind = 'note' | 'card' | 'text' | 'rectangle' | 'ellipse' | 'arrow'
export type BoardColor = 'sand' | 'blue' | 'green' | 'rose' | 'slate'
export type ReferenceKind = 'idea' | 'game' | 'site'
export type BoardGeometry = { x: number; y: number; width: number; height: number; z_index: number }
export type BoardNode = BoardGeometry & { id: number; kind: BoardKind; title: string; body: string; color: BoardColor; rotation: number; reference_kind: ReferenceKind | null; reference_id: number | null; version: number }
export type BoardReference = { kind: ReferenceKind; id: number; title: string; status: string; content_kind: string; missing: boolean }
export type BoardEdge = { id: number; source_id: number; target_id: number; source_handle: string; target_handle: string; routing: 'curve' | 'step'; label: string; color: BoardColor; arrow: boolean; version: number }
export type BoardDocument = { nodes: BoardNode[]; edges: BoardEdge[]; references: BoardReference[] }
export type BoardNodeInput = Omit<BoardNode, 'id' | 'version' | 'reference_kind' | 'reference_id'> & { version?: number; reference_kind: ReferenceKind | ''; reference_id: number }
export type BoardLayout = BoardGeometry & { id: number; version: number }
export type CanvasNode = Node<{ record: BoardNode; reference?: BoardReference }, 'canvas'>
export type CanvasEdge = Edge<{ record: BoardEdge }, 'relationship'>
export const boardKinds: Record<BoardKind, string> = { note: '便签', card: '内容卡片', text: '文本', rectangle: '矩形', ellipse: '圆形', arrow: '箭头' }
export const boardColors = [{ value: 'sand', label: '暖黄' }, { value: 'blue', label: '蓝色' }, { value: 'green', label: '绿色' }, { value: 'rose', label: '玫瑰' }, { value: 'slate', label: '灰色' }]
export const colorValues: Record<BoardColor, string> = { sand: '#be923e', blue: '#598fce', green: '#4ba780', rose: '#c5738b', slate: '#8291a6' }
export function nodeInput(node: BoardNode): BoardNodeInput {
  return { kind: node.kind, title: node.title, body: node.body, color: node.color, rotation: node.rotation, reference_kind: node.reference_kind ?? '', reference_id: node.reference_id ?? 0, version: node.version, x: node.x, y: node.y, width: node.width, height: node.height, z_index: node.z_index }
}
export function newNode(kind: BoardKind, position: { x: number; y: number }, z = 0): BoardNodeInput {
  return { kind, title: ['note', 'card', 'text'].includes(kind) ? `新${boardKinds[kind]}` : '', body: '', color: kind === 'note' ? 'sand' : 'slate', rotation: 0, reference_kind: '', reference_id: 0, x: Math.round(position.x), y: Math.round(position.y), width: kind === 'text' ? 240 : 300, height: kind === 'text' ? 80 : 220, z_index: z }
}
export function canvasNode(record: BoardNode, references: BoardReference[] = [], previous?: CanvasNode): CanvasNode {
  return { id: String(record.id), type: 'canvas', position: { x: record.x, y: record.y }, width: record.width, height: record.height, zIndex: record.z_index, selected: previous?.selected, data: { record, reference: references.find((item) => item.kind === record.reference_kind && item.id === record.reference_id) } }
}
export function mergeBoardNodes(document: BoardDocument, current: CanvasNode[], dirty: Set<string>): CanvasNode[] {
  const previous = new Map(current.map((node) => [node.id, node]))
  const next = document.nodes.map((row) => dirty.has(String(row.id)) && previous.has(String(row.id)) ? previous.get(String(row.id))! : canvasNode(row, document.references, previous.get(String(row.id))))
  // Keep a locally edited node even if another member deleted it; the conflict
  // remains visible until the member explicitly reloads/discards their draft.
  for (const node of current) if (dirty.has(node.id) && !next.some((item) => item.id === node.id)) next.push(node)
  return next
}
export function layoutInput(node: CanvasNode, overrides: Partial<BoardGeometry> = {}): BoardLayout {
  const round = Math.round
  return { id: node.data.record.id, version: node.data.record.version, x: round(node.position.x), y: round(node.position.y), width: round(node.width ?? node.data.record.width), height: round(node.height ?? node.data.record.height), z_index: node.zIndex ?? 0, ...overrides }
}
