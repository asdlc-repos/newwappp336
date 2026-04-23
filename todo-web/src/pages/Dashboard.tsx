import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Category,
  Task,
  createCategory,
  createTask,
  deleteCategory,
  deleteTask,
  listCategories,
  listTasks,
  toggleTaskComplete,
  updateTask,
} from '../api';
import { clearAuth } from '../auth';
import EditTaskModal from '../components/EditTaskModal';

type Filter = { kind: 'all' } | { kind: 'today' } | { kind: 'category'; id: string };

export default function Dashboard() {
  const navigate = useNavigate();
  const [categories, setCategories] = useState<Category[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [filter, setFilter] = useState<Filter>({ kind: 'all' });
  const [newCategoryName, setNewCategoryName] = useState('');
  const [newTaskTitle, setNewTaskTitle] = useState('');
  const [newTaskDueDate, setNewTaskDueDate] = useState('');
  const [newTaskCategoryId, setNewTaskCategoryId] = useState('');
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshCategories = useCallback(async () => {
    try {
      const cats = await listCategories();
      setCategories(cats);
    } catch (err) {
      setError((err as Error).message);
    }
  }, []);

  const refreshTasks = useCallback(async () => {
    try {
      const data = await listTasks();
      setTasks(data);
    } catch (err) {
      setError((err as Error).message);
    }
  }, []);

  useEffect(() => {
    (async () => {
      setLoading(true);
      await Promise.all([refreshCategories(), refreshTasks()]);
      setLoading(false);
    })();
  }, [refreshCategories, refreshTasks]);

  const filteredTasks = useMemo(() => {
    const now = new Date();
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const todayEnd = new Date(todayStart.getTime() + 24 * 60 * 60 * 1000);
    return tasks
      .filter((t) => {
        if (filter.kind === 'all') return true;
        if (filter.kind === 'category') return t.categoryId === filter.id;
        if (filter.kind === 'today') {
          if (!t.dueDate) return false;
          const d = new Date(t.dueDate);
          return d >= todayStart && d < todayEnd;
        }
        return true;
      })
      .sort((a, b) => {
        if (a.completed !== b.completed) return a.completed ? 1 : -1;
        const aDue = a.dueDate ? new Date(a.dueDate).getTime() : Infinity;
        const bDue = b.dueDate ? new Date(b.dueDate).getTime() : Infinity;
        if (aDue !== bDue) return aDue - bDue;
        return new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
      });
  }, [tasks, filter]);

  async function handleAddCategory(e: React.FormEvent) {
    e.preventDefault();
    const name = newCategoryName.trim();
    if (!name) return;
    try {
      const cat = await createCategory(name);
      setCategories((prev) => [...prev, cat]);
      setNewCategoryName('');
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function handleDeleteCategory(id: string) {
    if (!confirm('Delete this category? Tasks in it will become uncategorized.')) return;
    try {
      await deleteCategory(id);
      setCategories((prev) => prev.filter((c) => c.id !== id));
      if (filter.kind === 'category' && filter.id === id) {
        setFilter({ kind: 'all' });
      }
      // Refresh tasks since API nulls categoryId for them.
      await refreshTasks();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function handleAddTask(e: React.FormEvent) {
    e.preventDefault();
    const title = newTaskTitle.trim();
    if (!title) return;
    try {
      const task = await createTask({
        title,
        dueDate: newTaskDueDate ? new Date(newTaskDueDate).toISOString() : null,
        categoryId: newTaskCategoryId || null,
      });
      setTasks((prev) => [...prev, task]);
      setNewTaskTitle('');
      setNewTaskDueDate('');
      setNewTaskCategoryId('');
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function handleToggle(task: Task) {
    try {
      const updated = await toggleTaskComplete(task.id, !task.completed);
      setTasks((prev) => prev.map((t) => (t.id === task.id ? updated : t)));
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this task?')) return;
    try {
      await deleteTask(id);
      setTasks((prev) => prev.filter((t) => t.id !== id));
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function handleSaveEdit(task: Task) {
    try {
      const updated = await updateTask(task.id, {
        title: task.title,
        dueDate: task.dueDate,
        categoryId: task.categoryId,
        completed: task.completed,
      });
      setTasks((prev) => prev.map((t) => (t.id === task.id ? updated : t)));
      setEditingTask(null);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  function handleLogout() {
    clearAuth();
    navigate('/login', { replace: true });
  }

  const categoryById = useMemo(() => {
    const map: Record<string, Category> = {};
    for (const c of categories) map[c.id] = c;
    return map;
  }, [categories]);

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-dot" />
          <span>Todo</span>
        </div>

        <nav className="filters">
          <button
            className={'filter-btn' + (filter.kind === 'all' ? ' active' : '')}
            onClick={() => setFilter({ kind: 'all' })}
          >
            All tasks
          </button>
          <button
            className={'filter-btn' + (filter.kind === 'today' ? ' active' : '')}
            onClick={() => setFilter({ kind: 'today' })}
          >
            Today
          </button>
        </nav>

        <div className="section-label">Categories</div>
        <ul className="categories">
          {categories.length === 0 && <li className="muted small">No categories yet</li>}
          {categories.map((c) => (
            <li key={c.id}>
              <button
                className={
                  'filter-btn' +
                  (filter.kind === 'category' && filter.id === c.id ? ' active' : '')
                }
                onClick={() => setFilter({ kind: 'category', id: c.id })}
              >
                {c.name}
              </button>
              <button
                className="icon-btn"
                onClick={() => handleDeleteCategory(c.id)}
                aria-label={`Delete category ${c.name}`}
                title="Delete category"
              >
                ×
              </button>
            </li>
          ))}
        </ul>

        <form className="add-category" onSubmit={handleAddCategory}>
          <input
            type="text"
            placeholder="New category"
            value={newCategoryName}
            onChange={(e) => setNewCategoryName(e.target.value)}
          />
          <button type="submit" className="btn">
            Add
          </button>
        </form>

        <div className="sidebar-footer">
          <button className="btn ghost" onClick={handleLogout}>
            Log out
          </button>
        </div>
      </aside>

      <main className="main">
        <header className="main-header">
          <h1>
            {filter.kind === 'all' && 'All tasks'}
            {filter.kind === 'today' && 'Today'}
            {filter.kind === 'category' &&
              (categoryById[filter.id]?.name ?? 'Category')}
          </h1>
          <div className="muted">
            {filteredTasks.length} task{filteredTasks.length === 1 ? '' : 's'}
          </div>
        </header>

        {error && (
          <div className="alert" onClick={() => setError(null)}>
            {error} <span className="muted small">(click to dismiss)</span>
          </div>
        )}

        <form className="add-task" onSubmit={handleAddTask}>
          <input
            type="text"
            placeholder="What needs doing?"
            value={newTaskTitle}
            onChange={(e) => setNewTaskTitle(e.target.value)}
            required
          />
          <input
            type="datetime-local"
            value={newTaskDueDate}
            onChange={(e) => setNewTaskDueDate(e.target.value)}
          />
          <select
            value={newTaskCategoryId}
            onChange={(e) => setNewTaskCategoryId(e.target.value)}
          >
            <option value="">No category</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
          <button type="submit" className="btn primary">
            Add task
          </button>
        </form>

        {loading ? (
          <div className="muted">Loading…</div>
        ) : filteredTasks.length === 0 ? (
          <div className="empty">
            Nothing here yet. Add a task above to get started.
          </div>
        ) : (
          <ul className="task-list">
            {filteredTasks.map((t) => (
              <li key={t.id} className={'task-row' + (t.completed ? ' done' : '')}>
                <input
                  type="checkbox"
                  checked={t.completed}
                  onChange={() => handleToggle(t)}
                  aria-label={`Mark ${t.title} ${t.completed ? 'incomplete' : 'complete'}`}
                />
                <div className="task-main">
                  <div className="task-title">{t.title}</div>
                  <div className="task-meta">
                    {t.dueDate && (
                      <span className="badge">{formatDate(t.dueDate)}</span>
                    )}
                    {t.categoryId && categoryById[t.categoryId] && (
                      <span className="badge">{categoryById[t.categoryId].name}</span>
                    )}
                  </div>
                </div>
                <div className="task-actions">
                  <button className="btn ghost" onClick={() => setEditingTask(t)}>
                    Edit
                  </button>
                  <button className="btn ghost danger" onClick={() => handleDelete(t.id)}>
                    Delete
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </main>

      {editingTask && (
        <EditTaskModal
          task={editingTask}
          categories={categories}
          onClose={() => setEditingTask(null)}
          onSave={handleSaveEdit}
        />
      )}
    </div>
  );
}

function formatDate(iso: string): string {
  try {
    const d = new Date(iso);
    const now = new Date();
    const sameDay =
      d.getFullYear() === now.getFullYear() &&
      d.getMonth() === now.getMonth() &&
      d.getDate() === now.getDate();
    const dateStr = sameDay
      ? 'Today'
      : d.toLocaleDateString(undefined, {
          month: 'short',
          day: 'numeric',
        });
    const hasTime =
      d.getHours() !== 0 || d.getMinutes() !== 0 || d.getSeconds() !== 0;
    if (!hasTime) return dateStr;
    const timeStr = d.toLocaleTimeString(undefined, {
      hour: 'numeric',
      minute: '2-digit',
    });
    return `${dateStr} · ${timeStr}`;
  } catch {
    return iso;
  }
}
