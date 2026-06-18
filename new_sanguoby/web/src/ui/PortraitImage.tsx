import { useEffect, useState } from 'react';
import { DEFAULT_PORTRAIT } from '../game/portraitRegistry';

type PortraitImageProps = {
  src: string;
  alt: string;
  className: string;
  fallbackLabel?: string;
  loading?: 'eager' | 'lazy';
};

export function PortraitImage({ src, alt, className, fallbackLabel, loading = 'lazy' }: PortraitImageProps) {
  const [fallbackLevel, setFallbackLevel] = useState(0);

  useEffect(() => {
    setFallbackLevel(0);
  }, [src]);

  if (fallbackLevel > 1) {
    return (
      <span className={`${className} portrait-fallback`} role="img" aria-label={alt}>
        {fallbackLabel?.trim().slice(0, 1) || alt.trim().slice(0, 1) || '将'}
      </span>
    );
  }

  return (
    <img
      src={fallbackLevel === 0 ? src : DEFAULT_PORTRAIT}
      alt={alt}
      className={className}
      decoding="async"
      loading={loading}
      onError={() => setFallbackLevel((level) => level + 1)}
    />
  );
}
