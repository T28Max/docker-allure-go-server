import React, { useEffect, useState } from 'react'

function fmtDate(s) { return new Date(s).toLocaleString() }

export default function App(){
  const [reports, setReports] = useState([])
  const [project, setProject] = useState('demo')
  const [file, setFile] = useState(null)
  const [loading, setLoading] = useState(false)
  const [msg, setMsg] = useState('')

  const load = async () => {
    const r = await fetch('/api/reports?project=' + encodeURIComponent(project))
    const j = await r.json()
    setReports(j)
  }

  useEffect(() => { load() }, [project])

  const onUpload = async (e) => {
    e.preventDefault()
    if (!file) return
    setLoading(true); setMsg('Uploading...')
    const fd = new FormData()
    fd.append('project', project)
    fd.append('file', file)
    const r = await fetch('/api/uploads', { method: 'POST', body: fd, headers: { 'Authorization': 'Bearer devtoken' } })
    if (!r.ok) { setMsg('Upload failed'); setLoading(false); return }
    setMsg('Processing... refresh in a bit.'); setFile(null)
    setTimeout(load, 2000)
    setLoading(false)
  }

  const onDelete = async (id) => {
    if (!confirm('Delete report?')) return
    const r = await fetch('/api/reports/' + id, { method: 'DELETE', headers: { 'Authorization': 'Bearer devtoken' } })
    if (r.ok) load()
  }

  return (
    <div className="min-h-screen">
      <header className="bg-white shadow p-4">
        <div className="max-w-5xl mx-auto flex items-center justify-between">
          <h1 className="text-xl font-semibold">Allure-Lite</h1>
          <div>
            <select value={project} onChange={e=>setProject(e.target.value)} className="border rounded px-2 py-1">
              <option value="demo">demo</option>
            </select>
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto p-4">
        <div className="grid md:grid-cols-3 gap-4 mb-6">
          <div className="md:col-span-2 bg-white rounded-xl shadow p-4">
            <h2 className="font-medium mb-2">Reports</h2>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-slate-500">
                    <th className="py-2">ID</th>
                    <th>Date</th>
                    <th>Status</th>
                    <th>Total</th>
                    <th>Passed</th>
                    <th>Failed</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  {reports.map(r => (
                    <tr key={r.id} className="border-t">
                      <td className="py-2">{r.id.slice(0,8)}</td>
                      <td>{fmtDate(r.created_at)}</td>
                      <td>{r.status}</td>
                      <td>{r.total}</td>
                      <td>{r.passed}</td>
                      <td>{r.failed}</td>
                      <td className="text-right">
                        <a href={`/reports/${r.project}/${r.id}/`} target="_blank" className="text-blue-600 hover:underline mr-3">Open</a>
                        <button onClick={()=>onDelete(r.id)} className="text-red-600 hover:underline">Delete</button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow p-4">
            <h2 className="font-medium mb-2">Upload</h2>
            <form onSubmit={onUpload}>
              <input type="file" accept=".zip,.tar,.zst,.tar.zst" onChange={e=>setFile(e.target.files?.[0])} className="block w-full border rounded p-2 mb-2"/>
              <button disabled={!file||loading} className="bg-blue-600 text-white px-4 py-2 rounded disabled:opacity-50">Upload</button>
            </form>
            {msg && <p className="text-sm text-slate-600 mt-2">{msg}</p>}
            <p className="text-xs text-slate-500 mt-2">Tip: For huge artifacts, prefer CI → S3 multipart, then API completion.</p>
          </div>
        </div>
      </main>
    </div>
  )
}
