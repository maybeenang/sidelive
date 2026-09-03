export type Status =
  | 'idle'
  | 'connecting'
  | 'connected'
  | 'reconnecting'
  | 'offline'
  | 'disconnected'
  | 'failed';
export interface Stream {
  id: string;
  providerId: string;
  identity: { username: string; displayName: string };
  status: Status;
  error?: string;
}
export interface LiveEvent {
  id: string;
  streamId: string;
  provider: string;
  type: string;
  timestamp: string;
  user: { username?: string; displayName: string; avatarUrl?: string };
  payload: { text: string };
}
export interface Overlay {
  id: string;
  name: string;
  sources: { overlayId: string; streamId: string }[];
  bounds: { X: number; Y: number; Width: number; Height: number };
  appearance: { opacity: number; fontSize: number; maxComments: number };
  behavior: {
    alwaysOnTop: boolean;
    clickThrough: boolean;
    visible: boolean;
    locked: boolean;
  };
}
export interface Provider {
  ID: string;
  Name: string;
  Capabilities: { Comments: boolean };
}
export interface Snapshot {
  workspace: { streams: Stream[]; overlays: Overlay[] };
  providers: Provider[];
}
