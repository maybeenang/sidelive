import {
  Activity,
  ChevronRight,
  Eye,
  EyeOff,
  Lock,
  MessageCircle,
  MonitorUp,
  PanelTop,
  Play,
  Radio,
  Settings2,
  Square,
  Trash2,
  Unlock,
  Wifi,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { call, on } from './bridge';
import { useApp } from './store';
import type { LiveEvent, Overlay, Snapshot, Stream } from './types';

const statusLabel: Record<string, string> = {
  idle: 'Siap',
  connecting: 'Menghubungkan',
  connected: 'Terhubung',
  reconnecting: 'Menyambung ulang',
  offline: 'Offline sementara',
  disconnected: 'Terputus',
  failed: 'Gagal',
};
function Avatar({ name, url }: { name: string; url?: string }) {
  return url ? (
    <img className="avatar" src={url} alt="" />
  ) : (
    <span className="avatar fallback">{name.slice(0, 1).toUpperCase()}</span>
  );
}
function EventList({
  events,
  limit = 8,
}: {
  events: LiveEvent[];
  limit?: number;
}) {
  return (
    <div className="event-list">
      {events.slice(0, limit).map((event) => (
        <article className="comment" key={event.id}>
          <Avatar name={event.user.displayName} url={event.user.avatarUrl} />
          <div>
            <div className="comment-meta">
              <strong>{event.user.displayName}</strong>
              <span>
                {new Date(event.timestamp).toLocaleTimeString('id-ID', {
                  hour: '2-digit',
                  minute: '2-digit',
                })}
              </span>
            </div>
            <p>{event.payload.text}</p>
          </div>
        </article>
      ))}
      {events.length === 0 && (
        <div className="empty">
          <MessageCircle size={22} />
          <span>Komentar live akan muncul di sini</span>
        </div>
      )}
    </div>
  );
}
function OverlayView() {
  const events = useApp((s) => s.events);
  const snapshot = useApp((s) => s.snapshot);
  const overlay = snapshot?.workspace.overlays[0];
  return (
    <main
      className={`overlay-view ${overlay?.behavior.locked ? 'locked' : ''}`}
      style={
        {
          '--overlay-opacity': overlay?.appearance.opacity ?? 0.86,
          '--comment-size': `${overlay?.appearance.fontSize ?? 16}px`,
        } as React.CSSProperties
      }
    >
      <header className="overlay-head">
        <span>
          <Radio size={14} /> LIVE COMMENTS
        </span>
        {!overlay?.behavior.locked && (
          <span className="drag">Drag untuk memindahkan</span>
        )}
      </header>
      <EventList events={events} limit={overlay?.appearance.maxComments} />
    </main>
  );
}
export function App() {
  const snapshot = useApp((s) => s.snapshot);
  const events = useApp((s) => s.events);
  const setSnapshot = useApp((s) => s.setSnapshot);
  const setStatus = useApp((s) => s.setStatus);
  const push = useApp((s) => s.push);
  const clear = useApp((s) => s.clear);
  const [username, setUsername] = useState('');
  const [provider, setProvider] = useState('tiktok');
  const [error, setError] = useState('');
  const [tab, setTab] = useState<'stream' | 'overlay'>('stream');
  const overlayMode =
    new URLSearchParams(location.search).get('view') === 'overlay';
  useEffect(() => {
    call<Snapshot>('Snapshot')
      .then(setSnapshot)
      .catch((e) => setError(String(e)));
    const offEvent = on('live:event', (v) => push(v as LiveEvent));
    const offStatus = on('stream:status', (v) => setStatus(v as Stream));
    return () => {
      offEvent();
      offStatus();
    };
  }, [push, setSnapshot, setStatus]);
  if (overlayMode) return <OverlayView />;
  if (!snapshot)
    return (
      <div className="loading">
        <Radio /> Menyiapkan SideLive…
      </div>
    );
  const stream = snapshot.workspace.streams[0];
  const overlay = snapshot.workspace.overlays[0];
  const refresh = () => call<Snapshot>('Snapshot').then(setSnapshot);
  const add = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await call('AddStream', provider, username);
      setUsername('');
      await refresh();
    } catch (err) {
      setError(String(err));
    }
  };
  const connect = async () => {
    if (!stream) return;
    setError('');
    try {
      stream.status === 'connected' || stream.status === 'connecting'
        ? await call('Disconnect', stream.id)
        : await call('Connect', stream.id);
    } catch (err) {
      setError(String(err));
    }
  };
  const remove = async () => {
    if (stream) {
      await call('RemoveStream', stream.id);
      clear();
      await refresh();
    }
  };
  const update = async (patch: Partial<Overlay>) => {
    const next = { ...overlay, ...patch };
    setSnapshot({
      ...snapshot,
      workspace: { ...snapshot.workspace, overlays: [next] },
    });
    try {
      await call('UpdateOverlay', next);
    } catch (err) {
      setError(String(err));
    }
  };
  return (
    <div className="shell">
      <aside>
        <div className="brand">
          <span className="logo">
            <Radio />
          </span>
          <div>
            <strong>SideLive</strong>
            <small>Audience companion</small>
          </div>
        </div>
        <nav>
          <button
            type="button"
            className={tab === 'stream' ? 'active' : ''}
            onClick={() => setTab('stream')}
          >
            <Activity />
            Live stream
          </button>
          <button
            type="button"
            className={tab === 'overlay' ? 'active' : ''}
            onClick={() => setTab('overlay')}
          >
            <PanelTop />
            Overlay
          </button>
        </nav>
        <div className="aside-foot">
          <span className="privacy">
            <Lock />
            100% lokal
          </span>
          <small>Tanpa akun. Tanpa cloud.</small>
        </div>
      </aside>
      <main className="content">
        <header className="topbar">
          <div>
            <h1>{tab === 'stream' ? 'Live stream' : 'Overlay'}</h1>
            <p>
              {tab === 'stream'
                ? 'Hubungkan akun dan pantau percakapan secara real-time.'
                : 'Atur tampilan komentar di atas aplikasi lain.'}
            </p>
          </div>
          <span className="version">MVP · v0.1</span>
        </header>
        {error && (
          <div className="error">
            {error}
            <button type="button" onClick={() => setError('')}>
              ×
            </button>
          </div>
        )}
        {tab === 'stream' ? (
          <>
            <section className="hero">
              <div>
                <span className="eyebrow">
                  <Wifi /> CONNECTION
                </span>
                <h2>
                  Keep your live audience
                  <br />
                  <em>in sight.</em>
                </h2>
                <p>
                  SideLive membawa komentar TikTok LIVE ke desktop tanpa
                  mengganggu layar utama Anda.
                </p>
              </div>
              <div className="orb">
                <Radio />
              </div>
            </section>
            {!stream ? (
              <section className="card connect-card">
                <div className="card-title">
                  <span className="icon purple">
                    <MonitorUp />
                  </span>
                  <div>
                    <h3>Tambahkan live stream</h3>
                    <p>MVP mendukung satu stream aktif.</p>
                  </div>
                </div>
                <form onSubmit={add}>
                  <label>
                    Platform
                    <select
                      value={provider}
                      onChange={(e) => setProvider(e.target.value)}
                    >
                      {snapshot.providers.map((p) => (
                        <option value={p.ID} key={p.ID}>
                          {p.Name}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label>
                    {provider === 'demo' ? 'Nama demo' : 'Username TikTok'}
                    <div className="input-prefix">
                      <span>@</span>
                      <input
                        required
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        placeholder={
                          provider === 'demo' ? 'creator_demo' : 'username'
                        }
                      />
                    </div>
                  </label>
                  <button className="primary" type="submit">
                    Tambahkan stream <ChevronRight />
                  </button>
                </form>
              </section>
            ) : (
              <section className="card stream-card">
                <div className="stream-user">
                  <span className="provider-mark">♪</span>
                  <div>
                    <h3>{stream.identity.displayName}</h3>
                    <p>
                      {stream.providerId === 'demo'
                        ? 'Demo stream'
                        : 'TikTok LIVE'}
                    </p>
                  </div>
                  <span className={`status ${stream.status}`}>
                    <i />
                    {statusLabel[stream.status]}
                  </span>
                </div>
                {stream.error && <p className="stream-error">{stream.error}</p>}
                <div className="stream-actions">
                  <button
                    type="button"
                    className={
                      stream.status === 'connected' ? 'stop' : 'primary'
                    }
                    onClick={connect}
                  >
                    {stream.status === 'connected' ? <Square /> : <Play />}
                    {stream.status === 'connected' ? 'Putuskan' : 'Hubungkan'}
                  </button>
                  <button
                    type="button"
                    className="icon-btn danger"
                    title="Hapus stream"
                    onClick={remove}
                  >
                    <Trash2 />
                  </button>
                </div>
              </section>
            )}
            <section className="activity card">
              <div className="section-head">
                <div>
                  <h3>Aktivitas terbaru</h3>
                  <p>
                    {events.length
                      ? `${events.length} komentar diterima sesi ini`
                      : 'Menunggu komentar pertama'}
                  </p>
                </div>
                <span className="live-pill">
                  <i /> LIVE FEED
                </span>
              </div>
              <EventList events={events} limit={5} />
            </section>
          </>
        ) : (
          <OverlayPanel overlay={overlay} events={events} update={update} />
        )}
      </main>
    </div>
  );
}
function OverlayPanel({
  overlay,
  events,
  update,
}: {
  overlay: Overlay;
  events: LiveEvent[];
  update: (v: Partial<Overlay>) => void;
}) {
  const behavior = (v: Partial<Overlay['behavior']>) =>
    update({ behavior: { ...overlay.behavior, ...v } });
  const appearance = (v: Partial<Overlay['appearance']>) =>
    update({ appearance: { ...overlay.appearance, ...v } });
  return (
    <div className="overlay-grid">
      <section className="card preview-card">
        <div className="section-head">
          <div>
            <h3>Preview overlay</h3>
            <p>Tampilan komentar saat live</p>
          </div>
          <button
            type="button"
            className="ghost"
            onClick={() => behavior({ visible: !overlay.behavior.visible })}
          >
            {overlay.behavior.visible ? <EyeOff /> : <Eye />}
            {overlay.behavior.visible ? 'Sembunyikan' : 'Tampilkan'}
          </button>
        </div>
        <div
          className="preview"
          style={{
            opacity: overlay.appearance.opacity,
            fontSize: overlay.appearance.fontSize,
          }}
        >
          <div className="preview-label">
            <Radio /> LIVE COMMENTS
          </div>
          <EventList
            events={
              events.length
                ? events
                : [
                    {
                      id: 'sample',
                      streamId: '',
                      provider: 'tiktok',
                      type: 'comment',
                      timestamp: new Date().toISOString(),
                      user: { displayName: 'Ayu Streamer' },
                      payload: { text: 'Halo! Semangat live-nya 👋' },
                    },
                  ]
            }
            limit={overlay.appearance.maxComments}
          />
        </div>
      </section>
      <section className="card settings-card">
        <div className="card-title">
          <span className="icon purple">
            <Settings2 />
          </span>
          <div>
            <h3>Penampilan</h3>
            <p>Perubahan disimpan otomatis.</p>
          </div>
        </div>
        <label className="range-label">
          <span>
            Opacity <b>{Math.round(overlay.appearance.opacity * 100)}%</b>
          </span>
          <input
            type="range"
            min="20"
            max="100"
            value={overlay.appearance.opacity * 100}
            onChange={(e) =>
              appearance({ opacity: Number(e.target.value) / 100 })
            }
          />
        </label>
        <label className="range-label">
          <span>
            Ukuran teks <b>{overlay.appearance.fontSize}px</b>
          </span>
          <input
            type="range"
            min="12"
            max="32"
            value={overlay.appearance.fontSize}
            onChange={(e) => appearance({ fontSize: Number(e.target.value) })}
          />
        </label>
        <label>
          Jumlah komentar
          <select
            value={overlay.appearance.maxComments}
            onChange={(e) =>
              appearance({ maxComments: Number(e.target.value) })
            }
          >
            {[4, 6, 8, 10, 15].map((v) => (
              <option key={v}>{v}</option>
            ))}
          </select>
        </label>
        <div className="toggles">
          <button
            type="button"
            className={overlay.behavior.locked ? 'toggle active' : 'toggle'}
            onClick={() =>
              behavior({
                locked: !overlay.behavior.locked,
                clickThrough: !overlay.behavior.locked,
              })
            }
          >
            {overlay.behavior.locked ? <Lock /> : <Unlock />}
            <span>
              <b>
                {overlay.behavior.locked ? 'Overlay terkunci' : 'Mode edit'}
              </b>
              <small>
                {overlay.behavior.locked
                  ? 'Klik menembus overlay'
                  : 'Overlay dapat dipindahkan'}
              </small>
            </span>
            <i />
          </button>
          <button
            type="button"
            className={
              overlay.behavior.alwaysOnTop ? 'toggle active' : 'toggle'
            }
            onClick={() =>
              behavior({ alwaysOnTop: !overlay.behavior.alwaysOnTop })
            }
          >
            <PanelTop />
            <span>
              <b>Selalu di atas</b>
              <small>Tetap terlihat di aplikasi lain</small>
            </span>
            <i />
          </button>
        </div>
      </section>
    </div>
  );
}
