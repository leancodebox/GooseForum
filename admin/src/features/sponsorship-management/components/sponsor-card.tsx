import { useEffect, useRef, useState } from 'react';
import { combine } from '@atlaskit/pragmatic-drag-and-drop/combine';
import {
  draggable,
  dropTargetForElements,
} from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import {
  attachClosestEdge,
  extractClosestEdge,
} from '@atlaskit/pragmatic-drag-and-drop-hitbox/closest-edge';
import type { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/types';
import { GripVertical, HeartHandshake, MoreHorizontal, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { cn } from '@/lib/utils';
import { useBoardContext } from './board-context';
import { DropIndicator } from './drop-indicator';
import { SponsorItem, Sponsors } from '../data/schema';

interface SponsorCardProps {
  sponsor: SponsorItem;
  level: keyof Sponsors;
  index: number;
  onEdit: () => void;
  onDelete: () => void;
}

export function SponsorCard({ sponsor, level, index, onEdit, onDelete }: SponsorCardProps) {
  const { registerSponsor, instanceId } = useBoardContext();
  const ref = useRef<HTMLDivElement>(null);
  const [dragging, setDragging] = useState(false);
  const [closestEdge, setClosestEdge] = useState<Edge | null>(null);

  const sponsorId = `${level}-${index}-${sponsor.name}`;
  const isFeatured = level === 'level0';
  const isGold = level === 'level1';

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    return combine(
      registerSponsor(sponsorId, { element, actionMenuTrigger: element }),
      draggable({
        element,
        getInitialData: () => ({
          type: 'sponsor',
          sponsorId,
          level,
          index,
          instanceId,
        }),
        onDragStart: () => setDragging(true),
        onDrop: () => setDragging(false),
      }),
      dropTargetForElements({
        element,
        canDrop: ({ source }) =>
          source.data.instanceId === instanceId && source.data.type === 'sponsor',
        getIsSticky: () => true,
        getData: ({ input, element }) => {
          const data = {
            type: 'sponsor',
            sponsorId,
            level,
            index,
          };
        return attachClosestEdge(data, {
          input,
          element,
          allowedEdges: ['left', 'right'],
        });
      },
        onDragEnter: ({ self }) => setClosestEdge(extractClosestEdge(self.data)),
        onDrag: ({ self }) => setClosestEdge(extractClosestEdge(self.data)),
        onDragLeave: () => setClosestEdge(null),
        onDrop: () => setClosestEdge(null),
      })
    );
  }, [sponsorId, level, index, instanceId, registerSponsor, sponsor.name]);

  return (
    <div
      ref={ref}
      className={cn(
        'group relative rounded-lg border bg-background transition hover:border-primary/30 hover:bg-primary/[0.02]',
        isFeatured ? 'p-4' : isGold ? 'p-3' : 'p-2.5',
        dragging && 'opacity-50 grayscale'
      )}
    >
      {closestEdge && <DropIndicator edge={closestEdge} />}

      <GripVertical className='absolute left-1.5 top-1.5 h-4 w-4 cursor-grab text-muted-foreground/30 opacity-0 transition-opacity group-hover:opacity-100' />

      <div className='flex items-start gap-3'>
        <Avatar className={cn('shrink-0 rounded-md border border-border/70 bg-muted', isFeatured ? 'h-12 w-12' : 'h-11 w-11')}>
          <AvatarImage src={sponsor.avatarUrl} alt={sponsor.name} className='object-cover' />
          <AvatarFallback className='rounded-md bg-muted text-muted-foreground'>
            <HeartHandshake className='h-5 w-5' />
          </AvatarFallback>
        </Avatar>

        <div className='min-w-0 flex-1'>
          <div className='flex min-w-0 items-center gap-2'>
            <h4 className='truncate text-sm font-semibold text-foreground group-hover:text-primary'>
              {sponsor.name}
            </h4>
            {sponsor.link && (
              <a
                href={sponsor.link}
                target='_blank'
                rel='noopener noreferrer'
                className='shrink-0 text-muted-foreground/40 transition-colors hover:text-primary'
                aria-label='访问赞助商主页'
                onClick={(event) => event.stopPropagation()}
              >
                <ExternalLink className='h-3.5 w-3.5' />
              </a>
            )}
          </div>
          <p className='mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground'>
            {sponsor.message || '感谢支持 GooseForum。'}
          </p>
        </div>
      </div>

      <div className='absolute right-2 top-2 opacity-0 transition-opacity group-hover:opacity-100'>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant='ghost'
              size='icon'
              className='h-7 w-7 rounded-md bg-background/90 shadow-sm'
            >
              <MoreHorizontal className='h-4 w-4' />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align='end'>
            <DropdownMenuItem onClick={onEdit}>编辑赞助商</DropdownMenuItem>
            <DropdownMenuItem className='text-destructive' onClick={onDelete}>
              删除赞助商
            </DropdownMenuItem>
            {sponsor.link && (
              <DropdownMenuItem asChild>
                <a href={sponsor.link} target='_blank' rel='noopener noreferrer'>
                  <ExternalLink className='mr-2 h-4 w-4' />
                  访问主页
                </a>
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  );
}
