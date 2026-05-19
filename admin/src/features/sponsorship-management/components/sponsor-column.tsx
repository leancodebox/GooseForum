import { useEffect, useMemo, useRef, useState } from 'react';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import {
  dropTargetForElements,
} from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import {
  attachClosestEdge,
  extractClosestEdge,
} from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/types';
import { HeartHandshake, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { useBoardContext } from './board-context';
import { SponsorCard } from './sponsor-card';
import { DropIndicator } from './drop-indicator';
import { SponsorItem, Sponsors } from '../data/schema';

interface SponsorColumnProps {
  level: keyof Sponsors;
  title: string;
  sponsors: SponsorItem[];
  onAdd: () => void;
  onEdit: (index: number) => void;
  onDelete: (index: number) => void;
}

export function SponsorColumn({
  level,
  title,
  sponsors,
  onAdd,
  onEdit,
  onDelete,
}: SponsorColumnProps) {
  const { registerLevel, instanceId } = useBoardContext();
  const ref = useRef<HTMLDivElement>(null);
  const [closestEdge, setClosestEdge] = useState<Edge | null>(null);
  const levelLabel = useMemo(() => title.replace(/\s*\(Level\s*\d+\)/, ''), [title]);
  const levelCode = useMemo(() => title.match(/Level\s*\d+/)?.[0] ?? level, [level, title]);
  const tone = {
    level0: 'diamond',
    level1: 'gold',
    level2: 'silver',
    level3: 'bronze',
  }[level];
  const badgeClass = {
    diamond: 'bg-blue-50 text-blue-700',
    gold: 'bg-amber-50 text-amber-700',
    silver: 'bg-gray-100 text-gray-600',
    bronze: 'bg-rose-50 text-rose-700',
  }[tone];
  const gridClass = {
    diamond: 'grid-cols-1 md:grid-cols-2',
    gold: 'grid-cols-1 sm:grid-cols-2 2xl:grid-cols-3',
    silver: 'grid-cols-2 sm:grid-cols-3 xl:grid-cols-4',
    bronze: 'grid-cols-2 sm:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5',
  }[tone];

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    return combine(
      registerLevel(level, { element }),
      dropTargetForElements({
        element,
        canDrop: ({ source }) =>
          source.data.instanceId === instanceId && source.data.type === 'sponsor',
        getIsSticky: () => true,
        getData: ({ input, element }) => {
          const data = {
            type: 'level',
            level,
          };
          return attachClosestEdge(data, {
            input,
            element,
            allowedEdges: ['top', 'bottom'],
          });
        },
        onDragEnter: ({ self, source }) => {
          if (source.data.type === 'sponsor') {
            setClosestEdge(extractClosestEdge(self.data));
          }
        },
        onDrag: ({ self, source }) => {
          if (source.data.type === 'sponsor') {
            setClosestEdge(extractClosestEdge(self.data));
          }
        },
        onDragLeave: () => setClosestEdge(null),
        onDrop: () => setClosestEdge(null),
      })
    );
  }, [level, instanceId, registerLevel]);

  return (
    <div
      ref={ref}
      className='relative space-y-3'
    >
      {closestEdge && <DropIndicator edge={closestEdge} />}
      <div
        className='flex items-center justify-between gap-3 border-b border-border/70 pb-2'
      >
        <div className='flex min-w-0 items-center gap-2'>
          <span className={cn('rounded px-1.5 py-0.5 text-[11px] font-semibold', badgeClass)}>
            {levelLabel}
          </span>
          <span className='text-xs font-medium text-muted-foreground'>{levelCode}</span>
          <span className='rounded-full bg-muted px-2 py-0.5 text-[11px] font-semibold text-muted-foreground'>
            {sponsors.length}
          </span>
        </div>
        <Button
          variant='ghost'
          size='sm'
          className='h-8 shrink-0 gap-1.5 px-2 text-xs'
          onClick={onAdd}
        >
          <Plus className='h-3.5 w-3.5' />
          添加
        </Button>
      </div>

      <div
        className={cn(
          'min-h-24 rounded-lg border border-dashed border-transparent',
          sponsors.length === 0 && 'border-muted-foreground/20 bg-muted/20'
        )}
      >
        {sponsors.length === 0 ? (
          <button
            type='button'
            onClick={onAdd}
            className='flex h-28 w-full flex-col items-center justify-center rounded-lg text-sm text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground'
          >
            <HeartHandshake className='mb-2 h-5 w-5' />
            添加这个级别的第一个赞助商
          </button>
        ) : (
          <div className={cn('grid gap-3', gridClass)}>
            {sponsors.map((sponsor, index) => (
              <SponsorCard
                key={`${level}-${sponsor.name}-${index}`}
                sponsor={sponsor}
                level={level}
                index={index}
                onEdit={() => onEdit(index)}
                onDelete={() => onDelete(index)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
