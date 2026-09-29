import { useCallback, useEffect, useRef, useState } from "react";

/** Ignore superseded requests and keep successful data visible during refresh. */
export function useAsync<T>(key: string, loader: () => Promise<T>) {
  const loadRef = useRef(loader);
  loadRef.current = loader;
  const generation = useRef(0);
  const currentKey = useRef(key);
  currentKey.current = key;
  const [state, setState] = useState<{ key: string; data?: T; error?: string; loading: boolean }>({ key, loading: true });
  const refresh = useCallback(async () => {
    if (currentKey.current !== key) return;
    const request = ++generation.current;
    setState((old) => ({ key, data: old.key === key ? old.data : undefined, loading: true }));
    try {
      const data = await loadRef.current();
      if (request === generation.current) setState({ key, data, loading: false });
    } catch (error) {
      if (request === generation.current) setState((old) => ({ ...old, error: String(error), loading: false }));
    }
  }, [key]);
  useEffect(() => { void refresh(); return () => { generation.current++; }; }, [refresh]);
  return { ...(state.key === key ? state : { loading: true, data: undefined, error: undefined }), refresh };
}
