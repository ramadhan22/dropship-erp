/**
 * Structured logging utility for frontend applications
 * Provides consistent logging format and correlation ID tracking
 */

export const LogLevel = {
  DEBUG: 0,
  INFO: 1,
  WARN: 2,
  ERROR: 3,
  FATAL: 4,
} as const;

export type LogLevelType = typeof LogLevel[keyof typeof LogLevel];

export interface LogEntry {
  timestamp: Date;
  level: LogLevelType;
  message: string;
  component?: string;
  operation?: string;
  correlationId?: string;
  userId?: string;
  sessionId?: string;
  url?: string;
  userAgent?: string;
  fields?: Record<string, unknown>;
  error?: Error;
  stack?: string;
}

export interface LoggerConfig {
  level: LogLevelType;
  enableConsole: boolean;
  enableRemote: boolean;
  remoteEndpoint?: string;
  batchSize?: number;
  flushInterval?: number;
  maxRetries?: number;
  component?: string;
}

class Logger {
  private config: LoggerConfig;
  private logBuffer: LogEntry[] = [];
  private flushTimer: NodeJS.Timeout | null = null;
  private correlationId: string | null = null;
  private userId: string | null = null;
  private sessionId: string | null = null;

  constructor(config: LoggerConfig) {
    this.config = {
      batchSize: 10,
      flushInterval: 5000,
      maxRetries: 3,
      ...config,
    };
    
    // Initialize session ID
    this.sessionId = this.generateSessionId();
    
    // Start flush timer if remote logging is enabled
    if (this.config.enableRemote) {
      this.startFlushTimer();
    }
  }

  private generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private generateCorrelationId(): string {
    return `corr_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private shouldLog(level: LogLevelType): boolean {
    return level >= this.config.level;
  }

  private formatLogEntry(entry: LogEntry): string {
    const timestamp = entry.timestamp.toISOString();
    const levelNames = ['DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'];
    const level = levelNames[entry.level] || 'UNKNOWN';
    const component = entry.component || this.config.component || 'unknown';
    const operation = entry.operation || 'unknown';
    const correlationId = entry.correlationId || this.correlationId || 'none';
    
    let message = `${timestamp} [${level}] [${component}] [${operation}] [${correlationId}] ${entry.message}`;
    
    // Add fields
    if (entry.fields && Object.keys(entry.fields).length > 0) {
      const fieldStrings = Object.entries(entry.fields)
        .map(([key, value]) => `${key}=${JSON.stringify(value)}`)
        .join(' ');
      message += ` ${fieldStrings}`;
    }
    
    // Add error information
    if (entry.error) {
      message += ` error=${entry.error.message}`;
      if (entry.stack) {
        message += ` stack=${entry.stack}`;
      }
    }
    
    return message;
  }

  private async log(entry: LogEntry): Promise<void> {
    if (!this.shouldLog(entry.level)) {
      return;
    }

    // Set default values
    entry.timestamp = entry.timestamp || new Date();
    entry.component = entry.component || this.config.component;
    entry.correlationId = entry.correlationId || this.correlationId || undefined;
    entry.userId = entry.userId || this.userId || undefined;
    entry.sessionId = entry.sessionId || this.sessionId || undefined;
    entry.url = entry.url || window.location.href;
    entry.userAgent = entry.userAgent || navigator.userAgent;

    // Console logging
    if (this.config.enableConsole) {
      const formattedMessage = this.formatLogEntry(entry);
      
      switch (entry.level) {
        case LogLevel.DEBUG:
          console.debug(formattedMessage);
          break;
        case LogLevel.INFO:
          console.info(formattedMessage);
          break;
        case LogLevel.WARN:
          console.warn(formattedMessage);
          break;
        case LogLevel.ERROR:
        case LogLevel.FATAL:
          console.error(formattedMessage);
          break;
      }
    }

    // Remote logging
    if (this.config.enableRemote) {
      this.logBuffer.push(entry);
      
      if (this.logBuffer.length >= (this.config.batchSize || 10)) {
        await this.flush();
      }
    }
  }

  private startFlushTimer(): void {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
    }
    
    this.flushTimer = setInterval(() => {
      this.flush();
    }, this.config.flushInterval);
  }

  private async flush(): Promise<void> {
    if (this.logBuffer.length === 0) {
      return;
    }

    const logsToSend = [...this.logBuffer];
    this.logBuffer = [];

    if (!this.config.remoteEndpoint) {
      return;
    }

    try {
      const response = await fetch(this.config.remoteEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Correlation-ID': this.correlationId || '',
          'X-User-ID': this.userId || '',
          'X-Session-ID': this.sessionId || '',
        },
        body: JSON.stringify({
          logs: logsToSend,
          metadata: {
            userAgent: navigator.userAgent,
            url: window.location.href,
            timestamp: new Date().toISOString(),
          },
        }),
      });

      if (!response.ok) {
        console.error('Failed to send logs to remote endpoint:', response.status);
        // Re-add logs to buffer for retry
        this.logBuffer.unshift(...logsToSend);
      }
    } catch (error) {
      console.error('Error sending logs to remote endpoint:', error);
      // Re-add logs to buffer for retry
      this.logBuffer.unshift(...logsToSend);
    }
  }

  // Public API methods
  setCorrelationId(correlationId: string): void {
    this.correlationId = correlationId;
  }

  setUserId(userId: string): void {
    this.userId = userId;
  }

  getCorrelationId(): string | null {
    return this.correlationId;
  }

  newCorrelationId(): string {
    this.correlationId = this.generateCorrelationId();
    return this.correlationId;
  }

  debug(message: string, operation?: string, fields?: Record<string, unknown>): void {
    this.log({
      level: LogLevel.DEBUG,
      message,
      operation,
      fields,
      timestamp: new Date(),
    });
  }

  info(message: string, operation?: string, fields?: Record<string, unknown>): void {
    this.log({
      level: LogLevel.INFO,
      message,
      operation,
      fields,
      timestamp: new Date(),
    });
  }

  warn(message: string, operation?: string, fields?: Record<string, unknown>): void {
    this.log({
      level: LogLevel.WARN,
      message,
      operation,
      fields,
      timestamp: new Date(),
    });
  }

  error(message: string, error?: Error, operation?: string, fields?: Record<string, unknown>): void {
    this.log({
      level: LogLevel.ERROR,
      message,
      error,
      operation,
      fields,
      stack: error?.stack,
      timestamp: new Date(),
    });
  }

  fatal(message: string, error?: Error, operation?: string, fields?: Record<string, unknown>): void {
    this.log({
      level: LogLevel.FATAL,
      message,
      error,
      operation,
      fields,
      stack: error?.stack,
      timestamp: new Date(),
    });
  }

  // Performance monitoring
  startTimer(operation: string): PerformanceTimer {
    return new PerformanceTimer(this, operation);
  }

  // Cleanup
  destroy(): void {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
    }
    this.flush(); // Final flush
  }
}

class PerformanceTimer {
  private logger: Logger;
  private operation: string;
  private startTime: number;

  constructor(logger: Logger, operation: string) {
    this.logger = logger;
    this.operation = operation;
    this.startTime = performance.now();
  }

  finish(message: string, fields?: Record<string, unknown>): void {
    const duration = performance.now() - this.startTime;
    this.logger.info(message, this.operation, {
      ...fields,
      duration: `${duration.toFixed(2)}ms`,
    });
  }

  finishWithError(message: string, error: Error, fields?: Record<string, unknown>): void {
    const duration = performance.now() - this.startTime;
    this.logger.error(message, error, this.operation, {
      ...fields,
      duration: `${duration.toFixed(2)}ms`,
    });
  }
}

// Default logger instance
const defaultLogger = new Logger({
  level: LogLevel.INFO,
  enableConsole: true,
  enableRemote: false,
  component: 'dropship-erp-ui',
});

// Convenience functions
export const debug = (message: string, operation?: string, fields?: Record<string, unknown>) => 
  defaultLogger.debug(message, operation, fields);

export const info = (message: string, operation?: string, fields?: Record<string, unknown>) => 
  defaultLogger.info(message, operation, fields);

export const warn = (message: string, operation?: string, fields?: Record<string, unknown>) => 
  defaultLogger.warn(message, operation, fields);

export const error = (message: string, err?: Error, operation?: string, fields?: Record<string, unknown>) => 
  defaultLogger.error(message, err, operation, fields);

export const fatal = (message: string, err?: Error, operation?: string, fields?: Record<string, unknown>) => 
  defaultLogger.fatal(message, err, operation, fields);

export const startTimer = (operation: string) => defaultLogger.startTimer(operation);

export const setCorrelationId = (correlationId: string) => defaultLogger.setCorrelationId(correlationId);

export const setUserId = (userId: string) => defaultLogger.setUserId(userId);

export const getCorrelationId = () => defaultLogger.getCorrelationId();

export const newCorrelationId = () => defaultLogger.newCorrelationId();

export { Logger, PerformanceTimer };
export default defaultLogger;