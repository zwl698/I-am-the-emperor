import {useStrategyMusic} from '../audio/useStrategyMusic';

export function MusicToggle() {
  const {busy, enabled, supported, toggle} = useStrategyMusic({ autoStart: true });

  if (!supported) {
    return null;
  }

  return (
    <button
      type="button"
      className={`music-toggle ${enabled ? 'music-toggle--on' : ''}`}
      onClick={toggle}
      disabled={busy}
      aria-label={enabled ? '关闭原创背景音乐' : '开启原创背景音乐，或等待首次操作自动播放'}
      aria-pressed={enabled}
      title={enabled ? '关闭原创背景音乐' : '首次操作后自动播放，点此立即开启'}
    >
      <span aria-hidden="true">{enabled ? '乐' : '启'}</span>
    </button>
  );
}
