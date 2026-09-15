export interface EditorIdentity {
  editorId: string;
}

export function withEditorId<T extends object>(item: T): T & EditorIdentity {
  return {
    ...item,
    editorId: (item as Partial<EditorIdentity>).editorId || crypto.randomUUID(),
  };
}

export function withoutEditorId<T extends EditorIdentity>(
  item: T,
): Omit<T, "editorId"> {
  const { editorId: _editorId, ...value } = item;
  return value;
}
