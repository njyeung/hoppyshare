import { supabase } from './supabase'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

export async function makeAuthenticatedRequest(url: string, options: RequestInit = {}) {
  const { data: { session } } = await supabase.auth.getSession()
  
  if (!session?.access_token) {
    throw new Error('No authentication token available')
  }

  const headers = {
    'Authorization': `Bearer ${session.access_token}`,
    'Content-Type': 'application/json',
    ...options.headers,
  }

  return fetch(url, {
    ...options,
    headers,
  })
}

// Basic API functions
export async function apiGet(endpoint: string) {
  return makeAuthenticatedRequest(endpoint, { method: 'GET' })
}

export async function apiPost(endpoint: string, data: any) {
  return makeAuthenticatedRequest(endpoint, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function apiPut(endpoint: string, data: any) {
  return makeAuthenticatedRequest(endpoint, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function apiDelete(endpoint: string) {
  return makeAuthenticatedRequest(endpoint, { method: 'DELETE' })
}

// React Query hooks
export function useApiQuery<T>(
  queryKey: string[],
  endpoint: string,
  options?: any
) {
  return useQuery({
    queryKey,
    queryFn: async (): Promise<T> => {
      const response = await apiGet(endpoint)
      if (!response.ok) {
        throw new Error('API request failed')
      }
      return response.json()
    },
    ...options,
  })
}

export function useApiMutation<T, V>(
  endpoint: string,
  method: 'POST' | 'PUT' | 'DELETE' = 'POST',
  options?: any
) {
  const queryClient = useQueryClient()
  
  return useMutation({
    mutationFn: async (data: V): Promise<T> => {
      let response
      if (method === 'POST') {
        response = await apiPost(endpoint, data)
      } else if (method === 'PUT') {
        response = await apiPut(endpoint, data)
      } else {
        response = await apiDelete(endpoint)
      }
      
      if (!response.ok) {
        throw new Error('API request failed')
      }
      return response.json()
    },
    onSuccess: () => {
      // Invalidate and refetch related queries
      if (options?.invalidateQueries) {
        queryClient.invalidateQueries({ queryKey: options.invalidateQueries })
      }
    },
    ...options,
  })
}