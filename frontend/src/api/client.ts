import axios, { AxiosInstance, AxiosError } from 'axios'

// Use empty string for relative URLs (will use Vite proxy in dev, same origin in prod)
// Or use VITE_API_URL if explicitly set for production
const API_BASE_URL = import.meta.env.VITE_API_URL || ''

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.setupInterceptors()
  }

  private setupInterceptors() {
    // Request interceptor - add auth token
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    // Response interceptor - handle errors
    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        if (error.response?.status === 401) {
          // Unauthorized - clear token and redirect to login
          localStorage.removeItem('token')
          localStorage.removeItem('user')
          window.location.href = '/login'
        }
        return Promise.reject(error)
      }
    )
  }

  public get<T>(url: string, params?: any) {
    return this.client.get<T>(url, { params })
  }

  public post<T>(url: string, data?: any) {
    return this.client.post<T>(url, data)
  }

  public put<T>(url: string, data?: any) {
    return this.client.put<T>(url, data)
  }

  public delete<T>(url: string) {
    return this.client.delete<T>(url)
  }
}

export const apiClient = new ApiClient()
