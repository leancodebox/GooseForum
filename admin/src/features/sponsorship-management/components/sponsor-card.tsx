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
import { MoreHorizontal, Heart, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
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
      className={`relative group flex flex-col p-3 bg-white rounded-xl border shadow-sm transition-all hover:shadow-md ${
        dragging ? 'opacity-50 grayscale' : ''
      } border-transparent`}
    >
      {closestEdge && <DropIndicator edge={closestEdge} />}
      
      <div className="flex items-start justify-between mb-2">
        <Avatar className="h-10 w-10 rounded-lg border">
          <AvatarImage src={sponsor.avatarUrl} alt={sponsor.name} className="object-cover" />
          <AvatarFallback>
            <Heart className="h-5 w-5 text-muted-foreground" />
          </AvatarFallback>
        </Avatar>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity"
            >
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={onEdit}>编辑赞助商</DropdownMenuItem>
            <DropdownMenuItem className="text-destructive" onClick={onDelete}>
              删除赞助商
            </DropdownMenuItem>
            {sponsor.link && (
              <DropdownMenuItem asChild>
                <a href={sponsor.link} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="mr-2 h-4 w-4" />
                  访问主页
                </a>
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <div className="flex-1 min-w-0">
        <h4 className="text-sm font-semibold truncate text-foreground mb-1">
          {sponsor.name}
        </h4>
        <p className="text-xs text-muted-foreground line-clamp-2 mb-1.5 h-8">
          {sponsor.message}
        </p>
      </div>
    </div>
  );
}
