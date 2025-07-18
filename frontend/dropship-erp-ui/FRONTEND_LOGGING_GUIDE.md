# Frontend Logging Guide

This guide explains how to use the structured logging system in the frontend application.

## Overview

The frontend logging system provides:
- **Structured logging** with consistent format
- **Correlation ID tracking** across components and API calls
- **Performance monitoring** with automatic timing
- **Remote logging** to backend endpoints
- **React hooks** for easy integration

## Basic Usage

### Simple Logging

```typescript
import { info, error, warn, debug } from '../utils/logger';

// Basic info logging
info('User logged in successfully', 'UserAuth', { userId: '123' });

// Error logging with stack trace
try {
  // some operation
} catch (err) {
  error('Failed to process data', err as Error, 'DataProcessing', { 
    dataSize: data.length 
  });
}

// Warning logging
warn('API response slow', 'ApiCall', { 
  duration: '2.5s', 
  endpoint: '/api/data' 
});
```

### Component-Specific Logger

```typescript
import { Logger, LogLevel } from '../utils/logger';

const componentLogger = new Logger({
  level: LogLevel.INFO,
  enableConsole: true,
  enableRemote: true,
  remoteEndpoint: '/api/logs',
  component: 'UserProfileComponent',
});

// Use in component
componentLogger.info('Component mounted', 'ComponentMount', {
  userId: props.userId,
  timestamp: new Date().toISOString(),
});
```

## React Hooks

### useCorrelationId Hook

```typescript
import { useCorrelationId } from '../hooks/useCorrelationId';

function MyComponent() {
  const { correlationId, newCorrelationId, withCorrelationId } = useCorrelationId();

  const handleOperation = async () => {
    await withCorrelationId(async () => {
      // All logging within this function will use the same correlation ID
      info('Starting operation', 'HandleOperation');
      await performApiCall();
      info('Operation completed', 'HandleOperation');
    });
  };

  return (
    <div>
      <p>Correlation ID: {correlationId}</p>
      <button onClick={handleOperation}>Run Operation</button>
    </div>
  );
}
```

### useApiWithCorrelation Hook

```typescript
import { useApiWithCorrelation } from '../hooks/useCorrelationId';

function DataComponent() {
  const { fetchWithCorrelation } = useApiWithCorrelation();

  const fetchData = async () => {
    try {
      const response = await fetchWithCorrelation('/api/data', {
        method: 'GET',
      });
      const data = await response.json();
      // Automatically logged with correlation ID
      return data;
    } catch (error) {
      // Automatically logged with correlation ID
      throw error;
    }
  };

  return (
    <button onClick={fetchData}>Fetch Data</button>
  );
}
```

### usePerformanceMonitoring Hook

```typescript
import { usePerformanceMonitoring } from '../hooks/useCorrelationId';

function PerformanceComponent() {
  const { measurePerformance } = usePerformanceMonitoring();

  const heavyOperation = async () => {
    await measurePerformance('HeavyOperation', async () => {
      // Perform heavy computation
      await processLargeDataset();
    }, { dataSize: 1000000 });
  };

  return (
    <button onClick={heavyOperation}>Run Heavy Operation</button>
  );
}
```

## Error Boundaries

```typescript
import React from 'react';
import { useErrorBoundary } from '../hooks/useCorrelationId';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    const { reportError } = useErrorBoundary();
    reportError(error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return <h1>Something went wrong.</h1>;
    }

    return this.props.children;
  }
}
```

## Configuration

### Logger Configuration

```typescript
const loggerConfig = {
  level: LogLevel.INFO,           // Minimum log level
  enableConsole: true,            // Log to browser console
  enableRemote: true,             // Send logs to backend
  remoteEndpoint: '/api/logs',    // Backend endpoint for logs
  component: 'MyComponent',       // Component name for all logs
  batchSize: 10,                  // Number of logs to batch before sending
  flushInterval: 5000,            // Interval to flush logs (ms)
  maxRetries: 3,                  // Max retries for failed remote logging
};
```

### Environment-Specific Configuration

```typescript
// Development
const devLogger = new Logger({
  level: LogLevel.DEBUG,
  enableConsole: true,
  enableRemote: false,
  component: 'DevComponent',
});

// Production
const prodLogger = new Logger({
  level: LogLevel.WARN,
  enableConsole: false,
  enableRemote: true,
  remoteEndpoint: '/api/logs',
  component: 'ProdComponent',
});
```

## Best Practices

### 1. Use Consistent Operation Names

```typescript
// Good
info('User authentication started', 'UserAuth');
info('User authentication completed', 'UserAuth');

// Bad
info('User login', 'Login');
info('Auth done', 'Authentication');
```

### 2. Include Relevant Context

```typescript
// Good
info('API call completed', 'ApiCall', {
  endpoint: '/api/users',
  method: 'GET',
  statusCode: 200,
  duration: '245ms',
  userId: '123',
});

// Bad
info('API call done', 'ApiCall');
```

### 3. Use Appropriate Log Levels

```typescript
// Debug: Development information
debug('Component state updated', 'StateUpdate', { newState });

// Info: General information
info('User action completed', 'UserAction', { action: 'save' });

// Warn: Potentially problematic situations
warn('API response slow', 'ApiCall', { duration: '3s' });

// Error: Error conditions
error('Failed to save data', error, 'DataSave', { userId: '123' });
```

### 4. Performance Monitoring

```typescript
// Use performance monitoring for expensive operations
const timer = startTimer('DataProcessing');
try {
  const result = await processLargeDataset(data);
  timer.finish('Data processing completed', { 
    recordCount: result.length,
    status: 'success' 
  });
} catch (error) {
  timer.finishWithError('Data processing failed', error, {
    recordCount: data.length,
  });
}
```

## Log Format

The frontend logging system produces structured log messages:

```
2025-07-18T19:26:03.123Z [INFO] [ComponentName] [Operation] [correlation-id] [duration] Message key=value error=message
```

### Example Log Entries

```
2025-07-18T19:26:03.123Z [INFO] [UserProfile] [LoadProfile] [corr_1642523163123_abc123] User profile loaded successfully userId="123" loadTime="245ms"

2025-07-18T19:26:03.456Z [ERROR] [ApiClient] [FetchData] [corr_1642523163123_abc123] [2.5s] Failed to fetch user data error=Network timeout stack=Error: Network timeout at fetch...

2025-07-18T19:26:03.789Z [WARN] [DataGrid] [RenderRows] [corr_1642523163123_abc123] Large dataset detected rowCount=10000 renderTime="1.2s"
```

## Remote Logging

### Backend Endpoint

The frontend sends logs to the backend in batches:

```typescript
// Endpoint: POST /api/logs
// Headers: X-Correlation-ID, X-User-ID, X-Session-ID
// Body:
{
  "logs": [
    {
      "timestamp": "2025-07-18T19:26:03.123Z",
      "level": 1,
      "message": "User action completed",
      "component": "UserProfile",
      "operation": "SaveProfile",
      "correlationId": "corr_1642523163123_abc123",
      "userId": "123",
      "sessionId": "session_1642523163123_def456",
      "url": "https://app.example.com/profile",
      "userAgent": "Mozilla/5.0...",
      "fields": {
        "profileId": "123",
        "changeCount": 3
      }
    }
  ],
  "metadata": {
    "userAgent": "Mozilla/5.0...",
    "url": "https://app.example.com/profile",
    "timestamp": "2025-07-18T19:26:03.123Z"
  }
}
```

### Backend Handler Example

```go
// In your backend handler
func HandleFrontendLogs(c *gin.Context) {
    var request struct {
        Logs []struct {
            Timestamp     time.Time              `json:"timestamp"`
            Level         int                    `json:"level"`
            Message       string                 `json:"message"`
            Component     string                 `json:"component"`
            Operation     string                 `json:"operation"`
            CorrelationID string                 `json:"correlationId"`
            UserID        string                 `json:"userId"`
            SessionID     string                 `json:"sessionId"`
            URL           string                 `json:"url"`
            UserAgent     string                 `json:"userAgent"`
            Fields        map[string]interface{} `json:"fields"`
        } `json:"logs"`
    }
    
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // Process logs (store in database, forward to logging service, etc.)
    for _, log := range request.Logs {
        logger.Info(context.Background(), "FrontendLog", log.Message, 
            map[string]interface{}{
                "level":          log.Level,
                "component":      log.Component,
                "operation":      log.Operation,
                "correlation_id": log.CorrelationID,
                "user_id":        log.UserID,
                "session_id":     log.SessionID,
                "url":            log.URL,
                "user_agent":     log.UserAgent,
                "fields":         log.Fields,
            })
    }
    
    c.JSON(200, gin.H{"success": true})
}
```

## Testing

### Unit Tests

```typescript
import { Logger, LogLevel } from '../utils/logger';

describe('Component with logging', () => {
  let logger: Logger;

  beforeEach(() => {
    logger = new Logger({
      level: LogLevel.DEBUG,
      enableConsole: false,
      enableRemote: false,
      component: 'TestComponent',
    });
  });

  it('should log user actions', () => {
    const consoleSpy = jest.spyOn(console, 'info').mockImplementation();
    
    logger.info('Test action', 'TestOperation', { userId: '123' });
    
    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('Test action')
    );
    
    consoleSpy.mockRestore();
  });
});
```

### Integration Tests

```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { useCorrelationId } from '../hooks/useCorrelationId';

// Test correlation ID propagation
test('should maintain correlation ID across operations', async () => {
  const TestComponent = () => {
    const { correlationId, newCorrelationId } = useCorrelationId();
    
    return (
      <div>
        <span data-testid="correlation-id">{correlationId}</span>
        <button onClick={newCorrelationId}>New ID</button>
      </div>
    );
  };

  render(<TestComponent />);
  
  const initialId = screen.getByTestId('correlation-id').textContent;
  
  fireEvent.click(screen.getByText('New ID'));
  
  const newId = screen.getByTestId('correlation-id').textContent;
  
  expect(newId).not.toBe(initialId);
  expect(newId).toMatch(/^corr_\d+_[a-z0-9]+$/);
});
```

## Troubleshooting

### Common Issues

1. **Logs not appearing in console**: Check log level configuration
2. **Remote logging failing**: Verify endpoint URL and network connectivity
3. **Performance impact**: Adjust batch size and flush interval
4. **Correlation ID not propagating**: Ensure proper hook usage

### Debug Mode

```typescript
// Enable debug logging
const debugLogger = new Logger({
  level: LogLevel.DEBUG,
  enableConsole: true,
  enableRemote: false,
  component: 'DebugComponent',
});

// Log all state changes
debugLogger.debug('State changed', 'StateUpdate', { 
  oldState, 
  newState 
});
```

This comprehensive logging system ensures consistent, trackable, and debuggable frontend applications with full correlation to backend operations.