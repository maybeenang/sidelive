import type { LiveEvent, Overlay, Stream } from './types';

type Handler = (value: unknown) => void;
declare global {
  interface Window {
    wails?: {
      Call: {
        ByName: (name: string, ...args: unknown[]) => Promise<unknown>;
      };
      Events: { On: (event: string, handler: Handler) => () => void };
    };
  }
}
const demoOverlay: Overlay = {
  id: 'overlay_main',
  name: 'Live comments',
  sources: [],
  bounds: { X: 80, Y: 80, Width: 420, Height: 560 },
  appearance: { opacity: 0.86, fontSize: 16, maxComments: 8 },
  behavior: {
    alwaysOnTop: true,
    clickThrough: false,
    visible: false,
    locked: false,
  },
};
let demoStream: Stream | undefined;
const listeners = new Map<string, Set<Handler>>();
let timer: number | undefined;
const emit = (name: string, value: unknown) =>
  listeners.get(name)?.forEach((fn) => {
    fn(value);
  });
const demoNames = ['Ayu', 'Bima', 'Citra', 'Danu'];
const demoText = [
  'Halo! Semangat live-nya 👋',
  'Audio jernih banget',
  'Salam dari Bandung!',
  'Boleh jelaskan bagian tadi?',
];
const mockCall = async (name: string, args: unknown[]) => {
  if (name.endsWith('.Snapshot'))
    return {
      workspace: {
        streams: demoStream ? [demoStream] : [],
        overlays: [demoOverlay],
      },
      providers: [
        { ID: 'tiktok', Name: 'TikTok LIVE', Capabilities: { Comments: true } },
        { ID: 'demo', Name: 'Demo stream', Capabilities: { Comments: true } },
      ],
    };
  if (name.endsWith('.AddStream')) {
    demoStream = {
      id: `stream_${Date.now()}`,
      providerId: String(args[0]),
      identity: { username: String(args[1]), displayName: `@${args[1]}` },
      status: 'idle',
    };
    demoOverlay.sources = [
      { overlayId: demoOverlay.id, streamId: demoStream.id },
    ];
    return demoStream;
  }
  if (name.endsWith('.RemoveStream')) {
    demoStream = undefined;
    return;
  }
  if (name.endsWith('.Connect') && demoStream) {
    demoStream = { ...demoStream, status: 'connected' };
    emit('stream:status', demoStream);
    let i = 0;
    timer = window.setInterval(() => {
      if (!demoStream) return;
      const event: LiveEvent = {
        id: `demo_${Date.now()}`,
        streamId: demoStream.id,
        provider: demoStream.providerId,
        type: 'comment',
        timestamp: new Date().toISOString(),
        user: {
          displayName: demoNames[i % demoNames.length],
          username: 'demo_user',
        },
        payload: { text: demoText[i++ % demoText.length] },
      };
      emit('live:event', event);
    }, 1800);
    return;
  }
  if (name.endsWith('.Disconnect') && demoStream) {
    clearInterval(timer);
    demoStream = { ...demoStream, status: 'disconnected' };
    emit('stream:status', demoStream);
    return;
  }
  if (name.endsWith('.UpdateOverlay')) Object.assign(demoOverlay, args[0]);
  if (name.endsWith('.SetOverlayVisible'))
    demoOverlay.behavior.visible = Boolean(args[0]);
};
const service = 'github.com/sidelive/sidelive/internal/app.Service';
export const call = <T>(method: string, ...args: unknown[]): Promise<T> =>
  window.wails?.Call
    ? (window.wails.Call.ByName(`${service}.${method}`, ...args) as Promise<T>)
    : (mockCall(`${service}.${method}`, args) as Promise<T>);
export const on = (event: string, handler: Handler) => {
  if (window.wails?.Events) return window.wails.Events.On(event, handler);
  const set = listeners.get(event) ?? new Set();
  set.add(handler);
  listeners.set(event, set);
  return () => set.delete(handler);
};
