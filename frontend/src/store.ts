import { create } from 'zustand';
import type { LiveEvent, Snapshot, Stream } from './types';

interface State {
  snapshot?: Snapshot;
  events: LiveEvent[];
  setSnapshot: (v: Snapshot) => void;
  setStatus: (v: Stream) => void;
  push: (v: LiveEvent) => void;
  clear: () => void;
}
export const useApp = create<State>((set) => ({
  events: [],
  setSnapshot: (snapshot) => set({ snapshot }),
  setStatus: (stream) =>
    set((s) =>
      s.snapshot
        ? {
            snapshot: {
              ...s.snapshot,
              workspace: {
                ...s.snapshot.workspace,
                streams: s.snapshot.workspace.streams.map((v) =>
                  v.id === stream.id ? stream : v,
                ),
              },
            },
          }
        : s,
    ),
  push: (event) => set((s) => ({ events: [event, ...s.events].slice(0, 100) })),
  clear: () => set({ events: [] }),
}));
