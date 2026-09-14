// Shaka error codes in, plain-language copy out. No numbers or jargon reach the viewer.

export type ErrorKind =
  | 'expired'
  | 'network'
  | 'offline'
  | 'notFound'
  | 'unsupported'
  | 'drm'
  | 'generic';

export interface ClassifiedError {
  kind: ErrorKind;
  /** Stable string for the QoE beacon: `shaka-1001`, `expired`, `unsupported`. */
  code: string;
  /** i18n key for the message shown to the viewer. */
  messageKey: string;
  /** Whether a plain retry could plausibly work. */
  retryable: boolean;
}

const NETWORK = 1;
const MEDIA = 3;
const MANIFEST = 4;
const STREAMING = 5;
const DRM = 6;

const BAD_HTTP_STATUS = 1001;
const HTTP_ERROR = 1002;
const TIMEOUT = 1003;

export interface ShakaErrorLike {
  category?: number;
  code?: number;
  data?: unknown[];
}

export function classifyShakaError(err: ShakaErrorLike | undefined, online = true): ClassifiedError {
  if (!online) {
    return { kind: 'offline', code: 'offline', messageKey: 'offline', retryable: true };
  }
  const category = err?.category ?? 0;
  const code = err?.code ?? 0;
  const tag = `shaka-${code || 'unknown'}`;

  if (category === NETWORK) {
    const status = Number(err?.data?.[1]);
    // The origin answers 403 for a signature that has expired or never verified.
    if (status === 403) {
      return { kind: 'expired', code: 'playback_not_authorized', messageKey: 'errExpired', retryable: false };
    }
    if (status === 404) {
      return { kind: 'notFound', code: 'not_found', messageKey: 'errNotFound', retryable: false };
    }
    if (code === BAD_HTTP_STATUS || code === HTTP_ERROR || code === TIMEOUT) {
      return { kind: 'network', code: tag, messageKey: 'errNetwork', retryable: true };
    }
    return { kind: 'network', code: tag, messageKey: 'errNetwork', retryable: true };
  }

  if (category === DRM) {
    return { kind: 'drm', code: tag, messageKey: 'errDrm', retryable: false };
  }

  if (category === MANIFEST || category === MEDIA) {
    return { kind: 'unsupported', code: tag, messageKey: 'errUnsupported', retryable: false };
  }

  if (category === STREAMING) {
    return { kind: 'network', code: tag, messageKey: 'errNetwork', retryable: true };
  }

  return { kind: 'generic', code: tag, messageKey: 'errGeneric', retryable: true };
}

export function unsupportedBrowser(): ClassifiedError {
  return { kind: 'unsupported', code: 'browser_unsupported', messageKey: 'errUnsupported', retryable: false };
}

export function expiredSignature(): ClassifiedError {
  return { kind: 'expired', code: 'signature_expired', messageKey: 'errExpired', retryable: false };
}
