import type { ViewMode, TabFilter } from './types';
import { toDateKey } from './dates';

// Sync React state with URL query params for bookmarkable/shareable views.
// Keys: ?view=timeline&date=2026-07-15&tech=t1&tab=confirmed

export function urlToState(search: string) {
  const p = new URLSearchParams(search);
  return {
    view: (p.get('view') as ViewMode) || undefined,
    date: p.get('date') || undefined,
    tech: p.get('tech') || undefined,
    bay: p.get('bay') || undefined,
    tab: (p.get('tab') as TabFilter) || undefined,
  };
}

export function stateToUrl(view: ViewMode, date: Date, tech: string, bay: string, tab: TabFilter): string {
  const p = new URLSearchParams();
  p.set('view', view);
  p.set('date', toDateKey(date));
  if (tech) p.set('tech', tech);
  if (bay) p.set('bay', bay);
  if (tab !== 'confirmed') p.set('tab', tab);
  return '?' + p.toString();
}
