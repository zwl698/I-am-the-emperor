import {useCallback, useEffect, useState} from 'react';

type AudioContextConstructor = typeof AudioContext;

type AudioWindow = Window &
  typeof globalThis & {
    webkitAudioContext?: AudioContextConstructor;
  };

type StrategyMusicOptions = {
  autoStart?: boolean;
};

const STEP_LOOKAHEAD_SECONDS = 0.7;
const TICK_MS = 90;
const START_TIMEOUT_MS = 1500;
const TEMPO = 58;
const STEP_SECONDS = 60 / TEMPO / 2;
const MASTER_GAIN = 0.062;

const LEAD_SCALE = [220, 261.63, 293.66, 329.63, 392, 440, 523.25];
const LEAD_PATTERN = [2, -1, -1, -1, 4, -1, 3, -1, 1, -1, -1, -1, 2, -1, 0, -1, 3, -1, -1, -1, 5, -1, 4, -1, 2, -1, -1, -1, 1, -1, 0, -1];
const PLUCK_PATTERN = [0, -1, -1, 2, -1, -1, 3, -1, 1, -1, -1, 2, -1, -1, 4, -1];
const BASS_PATTERN = [55, 65.41, 73.42, 98];
const PAD_CHORDS = [
  [110, 146.83, 220],
  [98, 130.81, 196],
  [73.42, 110, 164.81],
  [82.41, 123.47, 196],
];

type MusicSnapshot = {
  autoMuted: boolean;
  busy: boolean;
  enabled: boolean;
};

const musicSubscribers = new Set<(snapshot: MusicSnapshot) => void>();
let sharedEngine: StrategyMusicEngine | null = null;
let sharedEnabled = false;
let sharedBusy = false;
let sharedAutoMuted = false;

export function useStrategyMusic({ autoStart = false }: StrategyMusicOptions = {}) {
  const [musicState, setMusicState] = useState<MusicSnapshot>(() => currentMusicSnapshot());
  const supported = isMusicSupported();

  useEffect(() => {
    musicSubscribers.add(setMusicState);
    return () => {
      musicSubscribers.delete(setMusicState);
    };
  }, []);

  const toggle = useCallback(async () => {
    if (!supported || sharedBusy) {
      return;
    }
    if (sharedEnabled) {
      stopSharedMusic(true);
      return;
    }
    sharedAutoMuted = false;
    publishMusicState();
    await startSharedMusic();
  }, [supported]);

  useEffect(() => {
    if (!autoStart || !supported || musicState.enabled || musicState.busy || musicState.autoMuted) {
      return undefined;
    }
    const handleFirstGesture = (event: Event) => {
      if (event.target instanceof Element && event.target.closest('.music-toggle')) {
        return;
      }
      void startSharedMusic();
    };
    window.addEventListener('pointerdown', handleFirstGesture, true);
    window.addEventListener('keydown', handleFirstGesture, true);
    return () => {
      window.removeEventListener('pointerdown', handleFirstGesture, true);
      window.removeEventListener('keydown', handleFirstGesture, true);
    };
  }, [autoStart, musicState.autoMuted, musicState.busy, musicState.enabled, supported]);

  return {busy: musicState.busy, enabled: musicState.enabled, supported, toggle};
}

export function useStrategyMusicAutoStart() {
  useStrategyMusic({ autoStart: true });
}

function currentMusicSnapshot(): MusicSnapshot {
  return {
    autoMuted: sharedAutoMuted,
    busy: sharedBusy,
    enabled: sharedEnabled,
  };
}

function publishMusicState() {
  const snapshot = currentMusicSnapshot();
  musicSubscribers.forEach((subscriber) => subscriber(snapshot));
}

async function startSharedMusic(): Promise<boolean> {
  if (!isMusicSupported() || sharedBusy || sharedEnabled) {
    return false;
  }
  sharedBusy = true;
  publishMusicState();
  let next: StrategyMusicEngine | null = null;
  try {
    next = new StrategyMusicEngine();
    await withTimeout(next.start(), START_TIMEOUT_MS);
    sharedEngine = next;
    sharedEnabled = true;
    sharedAutoMuted = false;
    return true;
  } catch {
    next?.stop();
    sharedEngine = null;
    sharedEnabled = false;
    return false;
  } finally {
    sharedBusy = false;
    publishMusicState();
  }
}

function stopSharedMusic(rememberMute: boolean) {
  sharedEngine?.stop();
  sharedEngine = null;
  sharedEnabled = false;
  if (rememberMute) {
    sharedAutoMuted = true;
  }
  publishMusicState();
}

function withTimeout<T>(task: Promise<T>, timeoutMs: number): Promise<T> {
  return new Promise((resolve, reject) => {
    const timer = window.setTimeout(() => reject(new Error('music start timeout')), timeoutMs);
    task.then(
      (value) => {
        window.clearTimeout(timer);
        resolve(value);
      },
      (error: unknown) => {
        window.clearTimeout(timer);
        reject(error);
      },
    );
  });
}

function isMusicSupported(): boolean {
  if (typeof window === 'undefined') {
    return false;
  }
  const audioWindow = window as AudioWindow;
  return Boolean(audioWindow.AudioContext || audioWindow.webkitAudioContext);
}

class StrategyMusicEngine {
  private context: AudioContext;
  private master: GainNode;
  private tone: BiquadFilterNode;
  private delay: DelayNode;
  private feedback: GainNode;
  private wet: GainNode;
  private timer = 0;
  private nextStepAt = 0;
  private step = 0;

  constructor() {
    const audioWindow = window as AudioWindow;
    const AudioContextClass = audioWindow.AudioContext || audioWindow.webkitAudioContext;
    if (!AudioContextClass) {
      throw new Error('Web Audio is not supported');
    }
    this.context = new AudioContextClass();
    this.master = this.context.createGain();
    this.master.gain.value = 0.0001;
    this.tone = this.context.createBiquadFilter();
    this.tone.type = 'lowpass';
    this.tone.frequency.value = 1350;
    this.tone.Q.value = 0.7;
    this.delay = this.context.createDelay(1.2);
    this.delay.delayTime.value = 0.34;
    this.feedback = this.context.createGain();
    this.feedback.gain.value = 0.18;
    this.wet = this.context.createGain();
    this.wet.gain.value = 0.16;
    this.master.connect(this.tone);
    this.master.connect(this.delay);
    this.delay.connect(this.feedback);
    this.feedback.connect(this.delay);
    this.delay.connect(this.wet);
    this.wet.connect(this.tone);
    this.tone.connect(this.context.destination);
  }

  async start() {
    await this.context.resume();
    const now = this.context.currentTime;
    this.master.gain.setValueAtTime(0.0001, now);
    this.master.gain.exponentialRampToValueAtTime(MASTER_GAIN, now + 0.8);
    this.nextStepAt = now + 0.08;
    this.schedule();
    this.timer = window.setInterval(() => this.schedule(), TICK_MS);
  }

  stop() {
    if (this.timer) {
      window.clearInterval(this.timer);
      this.timer = 0;
    }
    const now = this.context.currentTime;
    this.master.gain.cancelScheduledValues(now);
    this.master.gain.setValueAtTime(Math.max(this.master.gain.value, 0.0001), now);
    this.master.gain.exponentialRampToValueAtTime(0.0001, now + 0.35);
    window.setTimeout(() => {
      void this.context.close();
    }, 420);
  }

  private schedule() {
    while (this.nextStepAt < this.context.currentTime + STEP_LOOKAHEAD_SECONDS) {
      this.scheduleStep(this.step, this.nextStepAt);
      this.nextStepAt += STEP_SECONDS;
      this.step = (this.step + 1) % 32;
    }
  }

  private scheduleStep(step: number, time: number) {
    if (step % 16 === 0) {
      this.playPad(PAD_CHORDS[(step / 16) % PAD_CHORDS.length], time);
    }

    const leadIndex = LEAD_PATTERN[step % LEAD_PATTERN.length];
    if (leadIndex >= 0) {
      this.playLead(LEAD_SCALE[leadIndex], time, STEP_SECONDS * 2.8);
    }

    const pluckIndex = PLUCK_PATTERN[step % PLUCK_PATTERN.length];
    if (pluckIndex >= 0) {
      this.playPluck(LEAD_SCALE[pluckIndex] / 2, time + STEP_SECONDS * 0.08);
    }

    if (step % 8 === 0) {
      this.playBass(BASS_PATTERN[(step / 8) % BASS_PATTERN.length], time);
    }
    if (step % 16 === 12) {
      this.playDrum(time);
    }
  }

  private playLead(frequency: number, time: number, duration: number) {
    const filter = this.context.createBiquadFilter();
    filter.type = 'lowpass';
    filter.frequency.setValueAtTime(1080, time);
    filter.Q.setValueAtTime(1.5, time);
    filter.connect(this.master);

    const gain = this.context.createGain();
    gain.gain.setValueAtTime(0.0001, time);
    gain.gain.linearRampToValueAtTime(0.038, time + 0.12);
    gain.gain.exponentialRampToValueAtTime(0.0001, time + duration);
    gain.connect(filter);

    const main = this.context.createOscillator();
    main.type = 'triangle';
    main.frequency.setValueAtTime(frequency, time);
    main.detune.setValueAtTime(-3, time);
    main.connect(gain);
    main.start(time);
    main.stop(time + duration + 0.08);

    const breath = this.context.createOscillator();
    breath.type = 'sine';
    breath.frequency.setValueAtTime(frequency * 2.01, time);
    breath.detune.setValueAtTime(4, time);
    breath.connect(gain);
    breath.start(time);
    breath.stop(time + duration + 0.08);
  }

  private playPluck(frequency: number, time: number) {
    const gain = this.context.createGain();
    gain.gain.setValueAtTime(0.0001, time);
    gain.gain.linearRampToValueAtTime(0.034, time + 0.018);
    gain.gain.exponentialRampToValueAtTime(0.0001, time + 1.05);
    gain.connect(this.master);

    const oscillator = this.context.createOscillator();
    oscillator.type = 'sine';
    oscillator.frequency.setValueAtTime(frequency, time);
    oscillator.connect(gain);
    oscillator.start(time);
    oscillator.stop(time + 1.12);
  }

  private playBass(frequency: number, time: number) {
    const gain = this.context.createGain();
    gain.gain.setValueAtTime(0.0001, time);
    gain.gain.linearRampToValueAtTime(0.036, time + 0.08);
    gain.gain.exponentialRampToValueAtTime(0.0001, time + STEP_SECONDS * 6.5);
    gain.connect(this.master);

    const oscillator = this.context.createOscillator();
    oscillator.type = 'sine';
    oscillator.frequency.setValueAtTime(frequency, time);
    oscillator.connect(gain);
    oscillator.start(time);
    oscillator.stop(time + STEP_SECONDS * 6.8);
  }

  private playDrum(time: number) {
    const gain = this.context.createGain();
    gain.gain.setValueAtTime(0.0001, time);
    gain.gain.linearRampToValueAtTime(0.026, time + 0.018);
    gain.gain.exponentialRampToValueAtTime(0.0001, time + 0.34);
    gain.connect(this.master);

    const oscillator = this.context.createOscillator();
    oscillator.type = 'sine';
    oscillator.frequency.setValueAtTime(96, time);
    oscillator.frequency.exponentialRampToValueAtTime(42, time + 0.18);
    oscillator.connect(gain);
    oscillator.start(time);
    oscillator.stop(time + 0.36);
  }

  private playPad(frequencies: number[], time: number) {
    const duration = STEP_SECONDS * 14.5;
    for (const frequency of frequencies) {
      const gain = this.context.createGain();
      gain.gain.setValueAtTime(0.0001, time);
      gain.gain.linearRampToValueAtTime(0.012, time + 0.7);
      gain.gain.setValueAtTime(0.012, time + duration - 1.4);
      gain.gain.exponentialRampToValueAtTime(0.0001, time + duration);
      gain.connect(this.master);

      const oscillator = this.context.createOscillator();
      oscillator.type = 'sine';
      oscillator.frequency.setValueAtTime(frequency, time);
      oscillator.detune.setValueAtTime((frequency % 3) * 2 - 2, time);
      oscillator.connect(gain);
      oscillator.start(time);
      oscillator.stop(time + duration + 0.2);
    }
  }
}
