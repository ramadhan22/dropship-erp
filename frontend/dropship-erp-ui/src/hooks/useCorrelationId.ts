import { useEffect, useState, useCallback } from 'react';
import { getCorrelationId, setCorrelationId, newCorrelationId, info, error } from '../utils/logger';

interface UseCorrelationIdReturn {
  correlationId: string | null;
  setCorrelationId: (id: string) => void;
  newCorrelationId: () => string;
  withCorrelationId: <T>(fn: () => Promise<T>) => Promise<T>;
}

/**
 * Hook for managing correlation IDs in React components
 * Provides automatic correlation ID generation and management
 */
export const useCorrelationId = (): UseCorrelationIdReturn => {
  const [correlationId, setCorrelationIdState] = useState<string | null>(getCorrelationId());

  useEffect(() => {
    // Generate new correlation ID if none exists
    if (!correlationId) {
      const newId = newCorrelationId();
      setCorrelationIdState(newId);
      info('Generated new correlation ID', 'CorrelationIdGenerated', { correlationId: newId });
    }
  }, [correlationId]);

  const setCorrelationIdWrapper = useCallback((id: string) => {
    setCorrelationId(id);
    setCorrelationIdState(id);
    info('Set correlation ID', 'CorrelationIdSet', { correlationId: id });
  }, []);

  const newCorrelationIdWrapper = useCallback(() => {
    const newId = newCorrelationId();
    setCorrelationIdState(newId);
    info('Generated new correlation ID', 'CorrelationIdGenerated', { correlationId: newId });
    return newId;
  }, []);

  const withCorrelationId = useCallback(async <T>(fn: () => Promise<T>): Promise<T> => {
    const currentId = getCorrelationId();
    const functionId = newCorrelationId();
    
    try {
      setCorrelationId(functionId);
      const result = await fn();
      return result;
    } catch (err) {
      error('Function execution failed', err as Error, 'WithCorrelationId', { 
        functionCorrelationId: functionId,
        originalCorrelationId: currentId,
      });
      throw err;
    } finally {
      // Restore original correlation ID
      if (currentId) {
        setCorrelationId(currentId);
      }
    }
  }, []);

  return {
    correlationId,
    setCorrelationId: setCorrelationIdWrapper,
    newCorrelationId: newCorrelationIdWrapper,
    withCorrelationId,
  };
};

/**
 * Hook for API calls with automatic correlation ID injection
 */
export const useApiWithCorrelation = () => {
  const { correlationId, newCorrelationId: generateNewId } = useCorrelationId();

  const fetchWithCorrelation = useCallback(async (url: string, options: RequestInit = {}): Promise<Response> => {
    const requestId = generateNewId();
    
    const headers = {
      'X-Correlation-ID': requestId,
      'X-Request-ID': requestId,
      ...options.headers,
    };

    info('Making API request', 'ApiRequest', {
      url,
      method: options.method || 'GET',
      correlationId: requestId,
    });

    try {
      const response = await fetch(url, {
        ...options,
        headers,
      });

      info('API request completed', 'ApiResponse', {
        url,
        status: response.status,
        statusText: response.statusText,
        correlationId: requestId,
        responseCorrelationId: response.headers.get('X-Correlation-ID'),
      });

      return response;
    } catch (err) {
      error('API request failed', err as Error, 'ApiRequestError', {
        url,
        correlationId: requestId,
      });
      throw err;
    }
  }, [generateNewId]);

  return {
    fetchWithCorrelation,
    correlationId,
  };
};

/**
 * Hook for performance monitoring with correlation IDs
 */
export const usePerformanceMonitoring = () => {
  const { correlationId } = useCorrelationId();

  const measurePerformance = useCallback(async <T>(
    operation: string,
    fn: () => Promise<T>,
    additionalFields?: Record<string, unknown>
  ): Promise<T> => {
    const startTime = performance.now();
    const measurementId = newCorrelationId();
    
    info('Performance measurement started', 'PerformanceStart', {
      operation,
      correlationId: measurementId,
      parentCorrelationId: correlationId,
      ...additionalFields,
    });

    try {
      const result = await fn();
      const duration = performance.now() - startTime;
      
      info('Performance measurement completed', 'PerformanceEnd', {
        operation,
        duration: `${duration.toFixed(2)}ms`,
        correlationId: measurementId,
        parentCorrelationId: correlationId,
        ...additionalFields,
      });

      return result;
    } catch (err) {
      const duration = performance.now() - startTime;
      
      error('Performance measurement failed', err as Error, 'PerformanceError', {
        operation,
        duration: `${duration.toFixed(2)}ms`,
        correlationId: measurementId,
        parentCorrelationId: correlationId,
        ...additionalFields,
      });

      throw err;
    }
  }, [correlationId]);

  return {
    measurePerformance,
    correlationId,
  };
};

/**
 * Hook for error boundary with correlation ID tracking
 */
export const useErrorBoundary = () => {
  const { correlationId } = useCorrelationId();

  const reportError = useCallback((err: Error, errorInfo?: { componentStack?: string }) => {
    error('Error boundary caught error', err, 'ErrorBoundary', {
      correlationId,
      componentStack: errorInfo?.componentStack,
      errorMessage: err.message,
      stack: err.stack,
    });
  }, [correlationId]);

  return {
    reportError,
    correlationId,
  };
};