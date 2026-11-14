import { useLayoutEffect, useRef, type RefObject } from "react";

export function useReturnFocus(returnFocus: RefObject<HTMLElement | null> | undefined) {
  const opener = useRef<Element | null>(null);
  // React sets a ref to null before it removes the element, so this keeps the last element.
  const popup = useRef<HTMLElement | null>(null);

  // The layout effect cleanup runs before React removes the elements, so it finds the focus in a dialog that the UI removes with its trigger.
  useLayoutEffect(
    () => () => {
      const active = document.activeElement;
      if (!active || !(active === opener.current || popup.current?.contains(active))) return;
      queueMicrotask(() => returnFocus?.current?.focus());
    },
    [returnFocus],
  );

  const popupRef = (element: HTMLElement | null) => {
    if (element) popup.current = element;
  };
  // Base UI calls this when the dialog opens, before it moves the focus, so the active element is the opener.
  const rememberOpener = () => {
    opener.current = document.activeElement;
  };
  const finalFocus = (): HTMLElement | true => {
    const element = opener.current;
    if (element?.isConnected && !element.matches(":disabled")) return true;
    return returnFocus?.current ?? true;
  };
  return { popupRef, rememberOpener, finalFocus };
}
