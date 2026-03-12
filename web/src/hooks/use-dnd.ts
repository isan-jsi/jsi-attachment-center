import { useState, useCallback, type RefCallback } from "react";

interface UseDragOptions {
  type: string;
  item: string;
}

export function useDrag({ type, item }: UseDragOptions) {
  const dragRef: RefCallback<HTMLElement> = useCallback(
    (node) => {
      if (!node) return;
      node.draggable = true;
      node.ondragstart = (e) => {
        e.dataTransfer?.setData(type, item);
      };
    },
    [type, item],
  );

  return { dragRef };
}

interface UseDropOptions {
  accept: string;
  onDrop: (item: string) => void;
}

export function useDrop({ accept, onDrop }: UseDropOptions) {
  const [isOver, setIsOver] = useState(false);

  const dropRef: RefCallback<HTMLElement> = useCallback(
    (node) => {
      if (!node) return;
      node.ondragover = (e) => {
        e.preventDefault();
        setIsOver(true);
      };
      node.ondragleave = () => setIsOver(false);
      node.ondrop = (e) => {
        e.preventDefault();
        setIsOver(false);
        const data = e.dataTransfer?.getData(accept);
        if (data) onDrop(data);
      };
    },
    [accept, onDrop],
  );

  return { isOver, dropRef };
}
