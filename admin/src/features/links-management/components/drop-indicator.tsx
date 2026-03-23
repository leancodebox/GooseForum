import { Edge } from '@atlaskit/pragmatic-drag-and-drop-hitbox/types';

interface DropIndicatorProps {
  edge: Edge;
  className?: string;
}

export function DropIndicator({ edge, className = '' }: DropIndicatorProps) {
  const positionClasses = {
    top: 'top-0 left-0 right-0 h-0.5 -translate-y-1',
    bottom: 'bottom-0 left-0 right-0 h-0.5 translate-y-1',
    left: 'left-0 top-0 bottom-0 w-0.5 -translate-x-1',
    right: 'right-0 top-0 bottom-0 w-0.5 translate-x-1',
  };

  const circleClasses = {
    top: 'left-0 -translate-x-1/2 -translate-y-[3px]',
    bottom: 'left-0 -translate-x-1/2 -translate-y-[3px]',
    left: 'top-0 -translate-y-1/2 -translate-x-[3px]',
    right: 'top-0 -translate-y-1/2 -translate-x-[3px]',
  };

  return (
    <div
      className={`absolute z-10 bg-primary pointer-events-none ${positionClasses[edge]} ${className}`}
    >
      <div
        className={`absolute w-2 h-2 rounded-full border-2 border-primary bg-white ${circleClasses[edge]}`}
      />
    </div>
  );
}
