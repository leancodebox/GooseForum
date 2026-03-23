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
import { MoreHorizontal, User } from 'lucide-react';
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
import { UserSponsor } from '../data/schema';

interface UserCardProps {
  user: UserSponsor;
  index: number;
  onEdit: () => void;
  onDelete: () => void;
}

export function UserCard({ user, index, onEdit, onDelete }: UserCardProps) {
  const { registerUser, instanceId } = useBoardContext();
  const ref = useRef<HTMLDivElement>(null);
  const [dragging, setDragging] = useState(false);
  const [closestEdge, setClosestEdge] = useState<Edge | null>(null);

  const userId = `user-${index}-${user.name}`;

  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    return combine(
      registerUser(userId, { element, actionMenuTrigger: element }),
      draggable({
        element,
        getInitialData: () => ({
          type: 'user',
          userId,
          index,
          instanceId,
        }),
        onDragStart: () => setDragging(true),
        onDrop: () => setDragging(false),
      }),
      dropTargetForElements({
        element,
        canDrop: ({ source }) =>
          source.data.instanceId === instanceId && source.data.type === 'user',
        getIsSticky: () => true,
        getData: ({ input, element }) => {
          const data = {
            type: 'user',
            userId,
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
  }, [userId, index, instanceId, registerUser, user.name]);

  return (
    <div
      ref={ref}
      className={`relative group flex items-center gap-2 p-2 bg-white rounded-xl border shadow-sm transition-all hover:shadow-md ${
        dragging ? 'opacity-50 grayscale' : ''
      } border-transparent`}
    >
      {closestEdge && <DropIndicator edge={closestEdge} />}
      
      <Avatar className="h-8 w-8 rounded-full border">
        <AvatarImage src={user.logo} alt={user.name} className="object-cover" />
        <AvatarFallback>
          <User className="h-4 w-4 text-muted-foreground" />
        </AvatarFallback>
      </Avatar>

      <div className="flex-1 min-w-0">
        <h4 className="text-sm font-semibold truncate text-foreground">
          {user.name}
        </h4>
        <div className="flex items-center gap-2 text-[10px] text-muted-foreground mt-0.5">
          <span className="font-medium text-primary">¥{user.amount}</span>
          <span>•</span>
          <span>{user.time}</span>
        </div>
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
          <DropdownMenuItem onClick={onEdit}>编辑用户</DropdownMenuItem>
          <DropdownMenuItem className="text-destructive" onClick={onDelete}>
            删除用户
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
