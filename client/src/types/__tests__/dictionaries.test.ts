import { describe, it, expect } from 'vitest';
import type {
  RequestStatusCode,
  MediaTypeCode,
  SourceStatusCode,
} from '../dictionaries';

describe('Dictionaries types', () => {
  it('should have valid RequestStatusCode values', () => {
    const codes: RequestStatusCode[] = ['pending', 'processing', 'completed', 'failed', 'partial'];
    expect(codes).toHaveLength(5);
  });

  it('should have valid MediaTypeCode values', () => {
    const codes: MediaTypeCode[] = ['image', 'video_audio', 'audio', 'document', 'archive', 'other'];
    expect(codes).toHaveLength(6);
  });

  it('should have valid SourceStatusCode values', () => {
    const codes: SourceStatusCode[] = ['active', 'inactive', 'error', 'blocked'];
    expect(codes).toHaveLength(4);
  });
});
