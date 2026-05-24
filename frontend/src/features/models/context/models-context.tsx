import React, { createContext, useContext, useState, useCallback, useMemo } from 'react';
import { Model } from '../data/schema';

export interface PendingModelAssociation {
  channelId: number;
  modelId: string;
}

type DialogType =
  | 'create'
  | 'batchCreate'
  | 'edit'
  | 'delete'
  | 'archive'
  | 'association'
  | 'developerAssociation'
  | 'settings'
  | 'bulkEnable'
  | 'bulkDisable'
  | 'unassociated'
  | null;

interface ModelsContextType {
  open: DialogType;
  setOpen: (open: DialogType) => void;
  currentRow: Model | null;
  setCurrentRow: (row: Model | null) => void;
  currentDeveloper: string | null;
  setCurrentDeveloper: (developer: string | null) => void;
  pendingAssociation: PendingModelAssociation | null;
  setPendingAssociation: (association: PendingModelAssociation | null) => void;
  selectedModels: Model[];
  setSelectedModels: (models: Model[]) => void;
  resetRowSelection: (() => void) | null;
  setResetRowSelection: (fn: (() => void) | null) => void;
}

const ModelsContext = createContext<ModelsContextType | undefined>(undefined);

export function ModelsProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useState<DialogType>(null);
  const [currentRow, setCurrentRow] = useState<Model | null>(null);
  const [currentDeveloper, setCurrentDeveloper] = useState<string | null>(null);
  const [pendingAssociation, setPendingAssociation] = useState<PendingModelAssociation | null>(null);
  const [selectedModels, setSelectedModels] = useState<Model[]>([]);
  const [resetRowSelection, setResetRowSelection] = useState<(() => void) | null>(null);

  const handleSetOpen = useCallback((newOpen: DialogType) => {
    setOpen(newOpen);
  }, []);

  const handleSetCurrentRow = useCallback((row: Model | null) => {
    setCurrentRow(row);
  }, []);

  const handleSetCurrentDeveloper = useCallback((developer: string | null) => {
    setCurrentDeveloper(developer);
  }, []);

  const handleSetPendingAssociation = useCallback((association: PendingModelAssociation | null) => {
    setPendingAssociation(association);
  }, []);

  const handleSetSelectedModels = useCallback((models: Model[]) => {
    setSelectedModels(models);
  }, []);

  const handleSetResetRowSelection = useCallback((fn: (() => void) | null) => {
    setResetRowSelection(() => fn);
  }, []);

  const value = useMemo(
    () => ({
      open,
      setOpen: handleSetOpen,
      currentRow,
      setCurrentRow: handleSetCurrentRow,
      currentDeveloper,
      setCurrentDeveloper: handleSetCurrentDeveloper,
      pendingAssociation,
      setPendingAssociation: handleSetPendingAssociation,
      selectedModels,
      setSelectedModels: handleSetSelectedModels,
      resetRowSelection,
      setResetRowSelection: handleSetResetRowSelection,
    }),
    [
      open,
      handleSetOpen,
      currentRow,
      handleSetCurrentRow,
      currentDeveloper,
      handleSetCurrentDeveloper,
      pendingAssociation,
      handleSetPendingAssociation,
      selectedModels,
      handleSetSelectedModels,
      resetRowSelection,
      handleSetResetRowSelection,
    ]
  );

  return <ModelsContext.Provider value={value}>{children}</ModelsContext.Provider>;
}

export function useModels() {
  const context = useContext(ModelsContext);
  if (context === undefined) {
    throw new Error('useModels must be used within a ModelsProvider');
  }
  return context;
}

export default ModelsProvider;
