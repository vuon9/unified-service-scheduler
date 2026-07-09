import { useEffect, useState } from 'react';
import type { ToastMessage, ToastType } from '../types';
import styles from './Toast.module.css';

interface ToastContainerProps {
  toasts: ToastMessage[];
  removeToast: (id: number) => void;
}

export default function ToastContainer({ toasts, removeToast }: ToastContainerProps) {
  return (
    <div className={styles.container} aria-live="polite">
      {toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} onDismiss={() => removeToast(toast.id)} />
      ))}
    </div>
  );
}

function ToastItem({ toast, onDismiss }: { toast: ToastMessage; onDismiss: () => void }) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const showTimer = setTimeout(() => setVisible(true), 10);
    const hideTimer = setTimeout(() => {
      setVisible(false);
      setTimeout(onDismiss, 300);
    }, 4000);
    return () => {
      clearTimeout(showTimer);
      clearTimeout(hideTimer);
    };
  }, [onDismiss]);

  const icon = toast.type === 'success' ? '✓' : '✕';

  return (
    <div className={`${styles.toast} ${styles[toast.type]} ${visible ? styles.visible : ''}`}>
      <span className={styles.icon}>{icon}</span>
      <span className={styles.message}>{toast.message}</span>
    </div>
  );
}

export function createToast(type: ToastType, message: string): ToastMessage {
  return { id: Date.now(), type, message };
}
