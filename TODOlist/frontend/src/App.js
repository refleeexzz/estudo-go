import React, { useState, useEffect } from 'react';
import './App.css';

function App() {
  const [tasks, setTasks] = useState([]);
  const [text, setText] = useState('');
  const [description, setDescription] = useState('');
  const [date, setDate] = useState('');
  const [time, setTime] = useState('');
  const [loading, setLoading] = useState(false);
  const [filter, setFilter] = useState('all');
  const [editIdx, setEditIdx] = useState(null);
  const [editTask, setEditTask] = useState({ text: '', description: '', date: '', time: '' });
  const [theme, setTheme] = useState('light');

  useEffect(() => {
    fetch('/api/tasks')
      .then(res => res.json())
      .then(data => setTasks(data))
      .catch(() => setTasks([]));
  }, []);

  // Notificação simples de lembrete
  useEffect(() => {
    const interval = setInterval(() => {
      const now = new Date();
      tasks.forEach(task => {
        if (!task.done && task.date && task.time) {
          const taskDate = new Date(task.date + 'T' + task.time);
          if (
            taskDate.getTime() - now.getTime() < 60000 &&
            taskDate.getTime() - now.getTime() > 0
          ) {
            alert(`Lembrete: ${task.text} às ${task.time} em ${task.date}`);
          }
        }
      });
    }, 30000);
    return () => clearInterval(interval);
  }, [tasks]);

  const addTask = async () => {
    if (text.trim() === '') return;
    setLoading(true);
    const res = await fetch('/api/tasks', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text, description, date, time, done: false })
    });
    const newTask = await res.json();
    setTasks([...tasks, newTask]);
    setText('');
    setDescription('');
    setDate('');
    setTime('');
    setLoading(false);
  };

  const toggleTask = async idx => {
    const task = tasks[idx];
    setLoading(true);
    const res = await fetch(`/api/tasks/${task.ID}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ...task, done: !task.done })
    });
    const updated = await res.json();
    setTasks(tasks.map((t, i) => i === idx ? updated : t));
    setLoading(false);
  };

  const removeTask = async idx => {
    const task = tasks[idx];
    setLoading(true);
    await fetch(`/api/tasks/${task.ID}`, { method: 'DELETE' });
    setTasks(tasks.filter((_, i) => i !== idx));
    setLoading(false);
  };

  // Edição inline
  const startEdit = idx => {
    setEditIdx(idx);
    setEditTask({
      text: tasks[idx].text,
      description: tasks[idx].description,
      date: tasks[idx].date,
      time: tasks[idx].time
    });
  };
  const saveEdit = async idx => {
    setLoading(true);
    const task = tasks[idx];
    const res = await fetch(`/api/tasks/${task.ID}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ...task, ...editTask })
    });
    const updated = await res.json();
    setTasks(tasks.map((t, i) => i === idx ? updated : t));
    setEditIdx(null);
    setLoading(false);
  };
  const cancelEdit = () => setEditIdx(null);

  // Ordenação por data/hora
  const sortedTasks = [...tasks].sort((a, b) => {
    if (!a.date && !b.date) return 0;
    if (!a.date) return 1;
    if (!b.date) return -1;
    const dateA = new Date(a.date + 'T' + (a.time || '00:00'));
    const dateB = new Date(b.date + 'T' + (b.time || '00:00'));
    return dateA - dateB;
  });

  // Filtros
  const filteredTasks = sortedTasks.filter(task => {
    if (filter === 'all') return true;
    if (filter === 'done') return task.done;
    if (filter === 'pending') return !task.done;
    if (filter === 'today') {
      const today = new Date().toISOString().slice(0, 10);
      return task.date === today;
    }
    return true;
  });

  useEffect(() => {
    document.body.className = theme === 'dark' ? 'dark' : '';
  }, [theme]);

  return (
    <div className={`todo-container ${theme === 'dark' ? 'dark' : ''}`}>
      <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 32}}>
        <h1>Minha TODO List</h1>
        <button onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')} style={{padding: '8px 16px', borderRadius: 8, border: 'none', background: theme === 'dark' ? '#2d3748' : '#3182ce', color: '#fff', cursor: 'pointer'}}>Tema: {theme === 'light' ? 'Claro' : 'Escuro'}</button>
      </div>
      <div className="todo-input-wrapper" style={{flexDirection: 'column', gap: 16}}>
        <input
          type="text"
          value={text}
          onChange={e => setText(e.target.value)}
          placeholder="Título da tarefa"
          className="todo-input"
          disabled={loading}
        />
        <textarea
          value={description}
          onChange={e => setDescription(e.target.value)}
          placeholder="Descrição"
          className="todo-input"
          style={{resize: 'vertical', minHeight: 40}}
          disabled={loading}
        />
        <div style={{display: 'flex', gap: 8}}>
          <input
            type="date"
            value={date}
            onChange={e => setDate(e.target.value)}
            className="todo-input"
            style={{flex: 1}}
            disabled={loading}
          />
          <input
            type="time"
            value={time}
            onChange={e => setTime(e.target.value)}
            className="todo-input"
            style={{flex: 1}}
            disabled={loading}
          />
        </div>
        <button onClick={addTask} className="todo-add-btn" disabled={loading}>Adicionar</button>
      </div>
      <div style={{display: 'flex', gap: 8, margin: '24px 0'}}>
        <button onClick={() => setFilter('all')} style={{padding: '6px 14px', borderRadius: 6, border: 'none', background: filter === 'all' ? '#3182ce' : '#e2e8f0', color: filter === 'all' ? '#fff' : '#2d3748', cursor: 'pointer'}}>Todas</button>
        <button onClick={() => setFilter('pending')} style={{padding: '6px 14px', borderRadius: 6, border: 'none', background: filter === 'pending' ? '#3182ce' : '#e2e8f0', color: filter === 'pending' ? '#fff' : '#2d3748', cursor: 'pointer'}}>Pendentes</button>
        <button onClick={() => setFilter('done')} style={{padding: '6px 14px', borderRadius: 6, border: 'none', background: filter === 'done' ? '#3182ce' : '#e2e8f0', color: filter === 'done' ? '#fff' : '#2d3748', cursor: 'pointer'}}>Concluídas</button>
        <button onClick={() => setFilter('today')} style={{padding: '6px 14px', borderRadius: 6, border: 'none', background: filter === 'today' ? '#3182ce' : '#e2e8f0', color: filter === 'today' ? '#fff' : '#2d3748', cursor: 'pointer'}}>Hoje</button>
      </div>
      <ul className="todo-list">
        {filteredTasks.map((task, idx) => (
          <li key={task.ID} className={task.done ? 'done' : ''} style={{flexDirection: 'column', alignItems: 'flex-start', gap: 8, background: theme === 'dark' ? '#2d3748' : undefined, color: theme === 'dark' ? '#fff' : undefined}}>
            <div style={{display: 'flex', alignItems: 'center', width: '100%'}}>
              {editIdx === idx ? (
                <>
                  <input
                    type="text"
                    value={editTask.text}
                    onChange={e => setEditTask({...editTask, text: e.target.value})}
                    className="todo-input"
                    style={{fontWeight: 'bold', fontSize: 18, flex: 1}}
                  />
                  <button onClick={() => saveEdit(idx)} className="todo-add-btn" style={{marginLeft: 8}}>Salvar</button>
                  <button onClick={cancelEdit} className="todo-remove-btn" style={{marginLeft: 8}}>Cancelar</button>
                </>
              ) : (
                <>
                  <span onClick={() => toggleTask(idx)} style={{fontWeight: 'bold', fontSize: 18, flex: 1, cursor: 'pointer'}}>{task.text}</span>
                  <button onClick={() => startEdit(idx)} className="todo-add-btn" style={{marginLeft: 8}}>Editar</button>
                  <button onClick={() => removeTask(idx)} className="todo-remove-btn" style={{marginLeft: 8}} disabled={loading}>Remover</button>
                </>
              )}
            </div>
            {editIdx === idx ? (
              <>
                <textarea
                  value={editTask.description}
                  onChange={e => setEditTask({...editTask, description: e.target.value})}
                  className="todo-input"
                  style={{resize: 'vertical', minHeight: 40}}
                />
                <div style={{display: 'flex', gap: 8}}>
                  <input
                    type="date"
                    value={editTask.date}
                    onChange={e => setEditTask({...editTask, date: e.target.value})}
                    className="todo-input"
                    style={{flex: 1}}
                  />
                  <input
                    type="time"
                    value={editTask.time}
                    onChange={e => setEditTask({...editTask, time: e.target.value})}
                    className="todo-input"
                    style={{flex: 1}}
                  />
                </div>
              </>
            ) : (
              <>
                {task.description && <div style={{color: theme === 'dark' ? '#cbd5e1' : '#4a5568', fontSize: 15}}>{task.description}</div>}
                <div style={{display: 'flex', gap: 16, fontSize: 14, color: theme === 'dark' ? '#cbd5e1' : '#718096'}}>
                  {task.date && <span>Data: {task.date}</span>}
                  {task.time && <span>Hora: {task.time}</span>}
                </div>
              </>
            )}
          </li>
        ))}
      </ul>
      {loading && <div style={{textAlign:'center',marginTop:16}}>Carregando...</div>}
    </div>
  );
}

export default App;
