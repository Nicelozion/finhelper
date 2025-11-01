const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export interface ApiResponse<T> {
  data: T;
  requestId?: string;
}

export async function api<T>(
  endpoint: string,
  options?: RequestInit
): Promise<ApiResponse<T>> {
  const url = `${API_BASE_URL}${endpoint}`;
  const requestId = crypto.randomUUID();

  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        'X-Request-Id': requestId,
        ...options?.headers,
      },
    });

    if (!response.ok) {
      const errorText = await response.text();
      let errorData: any;
      try {
        errorData = JSON.parse(errorText);
      } catch {
        errorData = { message: errorText || 'Unknown error' };
      }

      const error = new Error(errorData.message || `HTTP ${response.status}`);
      (error as any).status = response.status;
      (error as any).requestId = response.headers.get('X-Request-Id') || requestId;
      throw error;
    }

    const data = await response.json();
    return {
      data,
      requestId: response.headers.get('X-Request-Id') || requestId,
    };
  } catch (error: any) {
    if (error.requestId) {
      throw error;
    }
    // Network error
    error.requestId = requestId;
    throw error;
  }
}

