import { ref } from 'vue';
import type { TrickplayCue } from '~/types/netflix';

export function useTrickplay() {
  const cues = ref<TrickplayCue[]>([]);

  async function loadTrickplayVtt(vttUrl: string): Promise<void> {
    if (!vttUrl) return;
    const rawText = await $fetch<string>(vttUrl, { responseType: 'text' });
    cues.value = parseWebVttSprites(rawText, vttUrl);
  }

  function findCueAtMillisecond(targetMs: number): TrickplayCue | null {
    const list = cues.value;
    let low = 0;
    let high = list.length - 1;
    while (low <= high) {
      const mid = (low + high) >>> 1;
      const cue = list[mid];
      if (!cue) break;
      if (targetMs >= cue.startMs && targetMs < cue.endMs) {
        return cue;
      }
      if (targetMs < cue.startMs) {
        high = mid - 1;
      } else {
        low = mid + 1;
      }
    }
    return null;
  }

  return { cues, loadTrickplayVtt, findCueAtMillisecond };
}

function parseWebVttSprites(vttContent: string, baseVttUrl: string): TrickplayCue[] {
  const blocks = vttContent.replace(/\r\n/g, '\n').split('\n\n');
  const parsed: TrickplayCue[] = [];
  const baseDir = baseVttUrl.substring(0, baseVttUrl.lastIndexOf('/') + 1);

  for (const block of blocks) {
    const lines = block.trim().split('\n');
    const arrowLineIdx = lines.findIndex((l) => l.includes('-->'));
    if (arrowLineIdx === -1 || arrowLineIdx + 1 >= lines.length) continue;

    const timeLine = lines[arrowLineIdx] ?? '';
    const spriteLine = lines[arrowLineIdx + 1] ?? '';
    const [startRaw, endRaw] = timeLine.split('-->').map((s) => s.trim());
    if (!startRaw || !endRaw || !spriteLine.includes('#xywh=')) continue;

    const [filePart, coordPart] = spriteLine.split('#xywh=');
    if (!filePart || !coordPart) continue;
    const coords = coordPart.split(',').map((n) => parseInt(n, 10));
    if (coords.length !== 4) continue;

    parsed.push({
      startMs: parseVttTimestampToMs(startRaw),
      endMs: parseVttTimestampToMs(endRaw),
      spriteUrl: filePart.startsWith('http') ? filePart : `${baseDir}${filePart}`,
      x: coords[0] ?? 0,
      y: coords[1] ?? 0,
      w: coords[2] ?? 320,
      h: coords[3] ?? 180
    });
  }
  return parsed;
}

function parseVttTimestampToMs(ts: string): number {
  const parts = ts.split(':');
  if (parts.length !== 3) return 0;
  const hours = parseInt(parts[0] ?? '0', 10);
  const minutes = parseInt(parts[1] ?? '0', 10);
  const [secStr, msStr] = (parts[2] ?? '0.0').split('.');
  const seconds = parseInt(secStr ?? '0', 10);
  const millis = parseInt((msStr ?? '0').padEnd(3, '0').slice(0, 3), 10);
  return hours * 3600000 + minutes * 60000 + seconds * 1000 + millis;
}

