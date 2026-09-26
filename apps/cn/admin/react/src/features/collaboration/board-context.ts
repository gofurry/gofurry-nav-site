import { createContext } from 'react'
import type { ResizeParams } from '@xyflow/react'
import type { BoardEdge, BoardNode, BoardReference } from './board-types'

type Actions = { writable: boolean; openReference: (ref: BoardReference) => void; editNode: (node: BoardNode) => void; editEdge: (edge: BoardEdge) => void; deleteNode: (node: BoardNode) => void; deleteEdge: (edge: BoardEdge) => void; raise: (id: string) => void; beginResize: (id: string) => void; resize: (id: string, params: ResizeParams) => void }
export const BoardActions = createContext<Actions | null>(null)
