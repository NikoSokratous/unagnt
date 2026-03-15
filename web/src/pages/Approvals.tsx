import React from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { CheckCircle, XCircle, Clock } from 'lucide-react'
import LoadingSpinner from '../components/LoadingSpinner'
import './Approvals.css'

interface ApprovalRequest {
  id: string
  tool: string
  input: Record<string, unknown>
  approvers: string[]
  run_id: string
  step_id: string
  created_at: string
  status: string
}

interface PendingResponse {
  pending: ApprovalRequest[]
}

const fetchPending = async (): Promise<PendingResponse> => {
  const response = await fetch('/v1/approvals/pending')
  if (!response.ok) {
    throw new Error('Failed to fetch pending approvals')
  }
  return response.json()
}

const Approvals: React.FC = () => {
  const queryClient = useQueryClient()
  const { data, isLoading, error } = useQuery({
    queryKey: ['approvals-pending'],
    queryFn: fetchPending,
    refetchInterval: 3000,
  })

  const approveMutation = useMutation({
    mutationFn: async (id: string) => {
      const res = await fetch(`/v1/approvals/${id}/approve`, { method: 'POST' })
      if (!res.ok) throw new Error('Failed to approve')
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['approvals-pending'] }),
  })

  const denyMutation = useMutation({
    mutationFn: async (id: string) => {
      const res = await fetch(`/v1/approvals/${id}/deny`, { method: 'POST' })
      if (!res.ok) throw new Error('Failed to deny')
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['approvals-pending'] }),
  })

  if (isLoading) {
    return (
      <div className="approvals-page">
        <LoadingSpinner message="Loading approvals..." />
      </div>
    )
  }

  if (error) {
    return (
      <div className="approvals-page error">
        Error loading approvals: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    )
  }

  const pending = data?.pending ?? []

  return (
    <div className="approvals-page">
      <header className="page-header">
        <h1>Approvals</h1>
        <p className="subtitle">Review and approve pending human-in-the-loop requests</p>
        <div className="header-stats">
          <div className="stat-card">
            <div className="stat-value">{pending.length}</div>
            <div className="stat-label">Pending</div>
          </div>
        </div>
      </header>

      <div className="approvals-list">
        {pending.length === 0 ? (
          <div className="empty-state">
            <Clock size={48} />
            <h2>No pending approvals</h2>
            <p>Approval requests will appear here when agents need human sign-off</p>
          </div>
        ) : (
          pending.map((req) => (
            <div key={req.id} className="approval-card">
              <div className="approval-header">
                <div>
                  <h3 className="approval-tool">{req.tool}</h3>
                  {req.run_id && (
                    <p className="approval-run-id">Run: {req.run_id}</p>
                  )}
                </div>
              </div>
              <div className="approval-body">
                <div className="approval-input">
                  <label>Input</label>
                  <pre>{JSON.stringify(req.input, null, 2)}</pre>
                </div>
                {req.approvers && req.approvers.length > 0 && (
                  <div className="approval-approvers">
                    <label>Approvers: {req.approvers.join(', ')}</label>
                  </div>
                )}
              </div>
              <div className="approval-actions">
                <button
                  className="btn-approve"
                  onClick={() => approveMutation.mutate(req.id)}
                  disabled={approveMutation.isPending || denyMutation.isPending}
                >
                  <CheckCircle size={18} />
                  Approve
                </button>
                <button
                  className="btn-deny"
                  onClick={() => denyMutation.mutate(req.id)}
                  disabled={approveMutation.isPending || denyMutation.isPending}
                >
                  <XCircle size={18} />
                  Deny
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

export default Approvals
