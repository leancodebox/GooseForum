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
import { MoreHorizontal, Globe, EyeOff, Eye } from 'lucide-react';
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
import { Link } from '../data/schema';

interface LinkCardProps {
  link: Link;
  groupIdx: number;
  linkIdx: number;
  onEdit: () => void;
  onDelete: () => void;
  onToggleStatus: () => void;
}

export function LinkCard({ link, groupIdx, linkIdx, onEdit, onDelete, onToggleStatus }: LinkCardProps) {
  const { registerLink, instanceId } = useBoardContext();
  const ref = useRef<HTMLDivElement>(null);
  const [dragging, setDragging] = useState(false);
  const [closestEdge, setClosestEdge] = useState<Edge | null>(null);

  const linkId = `${groupIdx}-${linkIdx}-${link.name}`;

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    return combine(
      registerLink(linkId, { element, actionMenuTrigger: element }),
      draggable({
        element,
        getInitialData: () => ({
          type: 'link',
          linkId,
          groupIdx,
          linkIdx,
          instanceId,
        }),
        onDragStart: () => setDragging(true),
        onDrop: () => setDragging(false),
      }),
      dropTargetForElements({
        element,
        canDrop: ({ source }) =>
          source.data.instanceId === instanceId && source.data.type === 'link',
        getIsSticky: () => true,
        getData: ({ input, element }) => {
          const data = {
            type: 'link',
            linkId,
            groupIdx,
            linkIdx,
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
          }, [linkId, groupIdx, linkIdx, instanceId, registerLink]);

          return (
             <div
               ref={ref}
               className={`relative group flex items-center gap-3 p-3 bg-white rounded-xl border shadow-sm transition-all hover:shadow-md ${
                 dragging ? 'opacity-50 grayscale' : ''
               } ${link.status === 0 ? 'opacity-60 bg-gray-50/50' : ''} border-transparent`}
             >
               {closestEdge && <DropIndicator edge={closestEdge} />}
        <div className="relative group/status">
          <Avatar className={`h-10 w-10 rounded-full border ${link.status === 0 ? 'grayscale' : ''}`}>
            <AvatarImage src={link.logoUrl} alt={link.name} className="object-cover" />
            <AvatarFallback>
              <Globe className="h-5 w-5 text-muted-foreground" />
            </AvatarFallback>
          </Avatar>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onToggleStatus();
            }}
            className={`absolute -bottom-1 -right-1 rounded-full p-0.5 shadow-sm border transition-all ${
              link.status === 0 
                ? 'bg-white hover:bg-gray-100' 
                : 'bg-primary text-white opacity-0 group-hover/status:opacity-100'
            }`}
            title={link.status === 0 ? '点击展示链接' : '点击隐藏链接'}
          >
            {link.status === 0 ? (
              <EyeOff className="h-3 w-3" />
            ) : (
              <Eye className="h-3 w-3" />
            )}
          </button>
        </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between gap-2">
          <h4 className={`text-sm font-semibold truncate ${link.status === 0 ? 'text-muted-foreground' : 'text-foreground'}`}>
            {link.name}
          </h4>
        </div>
        <p className="text-xs text-muted-foreground truncate">{link.desc || link.url}</p>
      </div>

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
          <DropdownMenuItem onClick={onEdit}>编辑链接</DropdownMenuItem>
          <DropdownMenuItem className="text-destructive" onClick={onDelete}>
            删除链接
          </DropdownMenuItem>
          <DropdownMenuItem asChild>
            <a href={link.url} target="_blank" rel="noopener noreferrer">
              访问链接
            </a>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

