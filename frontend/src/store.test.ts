import { beforeEach, describe, expect, it } from 'vitest';
import { useApp } from './store';
import type { LiveEvent } from './types';

const event = (id: string): LiveEvent => ({
  id,
  streamId: 'stream',
  provider: 'demo',
  type: 'comment',
  timestamp: new Date().toISOString(),
  user: { displayName: 'Tester' },
  payload: { text: id },
});
describe('live event store', () => {
  beforeEach(() => useApp.setState({ events: [] }));
  it('keeps newest event first and bounds history', () => {
    for (let i = 0; i < 105; i++) useApp.getState().push(event(String(i)));
    expect(useApp.getState().events).toHaveLength(100);
    expect(useApp.getState().events[0].id).toBe('104');
    expect(useApp.getState().events.at(-1)?.id).toBe('5');
  });
});
