import { Logger, LogLevel } from '../utils/logger';

// Mock fetch for testing
global.fetch = jest.fn();

describe('Logger', () => {
  let logger: Logger;

  beforeEach(() => {
    jest.clearAllMocks();
    logger = new Logger({
      level: LogLevel.DEBUG,
      enableConsole: true,
      enableRemote: false,
      component: 'test-component',
    });
  });

  afterEach(() => {
    logger.destroy();
  });

  describe('Basic logging', () => {
    it('should log debug messages', () => {
      const consoleSpy = jest.spyOn(console, 'debug').mockImplementation();
      
      logger.debug('Test debug message', 'TestOperation', { key: 'value' });
      
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('[DEBUG] [test-component] [TestOperation]')
      );
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('Test debug message')
      );
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('key="value"')
      );
      
      consoleSpy.mockRestore();
    });

    it('should log info messages', () => {
      const consoleSpy = jest.spyOn(console, 'info').mockImplementation();
      
      logger.info('Test info message', 'TestOperation');
      
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('[INFO] [test-component] [TestOperation]')
      );
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('Test info message')
      );
      
      consoleSpy.mockRestore();
    });

    it('should log error messages with stack trace', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      const testError = new Error('Test error');
      
      logger.error('Test error message', testError, 'TestOperation');
      
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('[ERROR] [test-component] [TestOperation]')
      );
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('Test error message')
      );
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('error=Test error')
      );
      
      consoleSpy.mockRestore();
    });
  });

  describe('Log level filtering', () => {
    it('should respect log level filtering', () => {
      const warnLogger = new Logger({
        level: LogLevel.WARN,
        enableConsole: true,
        enableRemote: false,
        component: 'test-component',
      });

      const debugSpy = jest.spyOn(console, 'debug').mockImplementation();
      const infoSpy = jest.spyOn(console, 'info').mockImplementation();
      const warnSpy = jest.spyOn(console, 'warn').mockImplementation();
      
      warnLogger.debug('Debug message');
      warnLogger.info('Info message');
      warnLogger.warn('Warn message');
      
      expect(debugSpy).not.toHaveBeenCalled();
      expect(infoSpy).not.toHaveBeenCalled();
      expect(warnSpy).toHaveBeenCalled();
      
      debugSpy.mockRestore();
      infoSpy.mockRestore();
      warnSpy.mockRestore();
      warnLogger.destroy();
    });
  });

  describe('Correlation ID', () => {
    it('should generate and use correlation IDs', () => {
      const correlationId = logger.newCorrelationId();
      
      expect(correlationId).toMatch(/^corr_\d+_[a-z0-9]+$/);
      expect(logger.getCorrelationId()).toBe(correlationId);
    });

    it('should include correlation ID in log messages', () => {
      const consoleSpy = jest.spyOn(console, 'info').mockImplementation();
      
      const correlationId = logger.newCorrelationId();
      logger.info('Test message', 'TestOperation');
      
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining(`[${correlationId}]`)
      );
      
      consoleSpy.mockRestore();
    });
  });

  describe('Performance monitoring', () => {
    it('should measure operation duration', () => {
      const consoleSpy = jest.spyOn(console, 'info').mockImplementation();
      
      const timer = logger.startTimer('TestOperation');
      timer.finish('Operation completed', { result: 'success' });
      
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('Operation completed')
      );
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('duration=')
      );
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('result="success"')
      );
      
      consoleSpy.mockRestore();
    });

    it('should measure operation duration with error', () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
      const testError = new Error('Test error');
      
      const timer = logger.startTimer('TestOperation');
      timer.finishWithError('Operation failed', testError);
      
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('Operation failed')
      );
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('duration=')
      );
      expect(consoleSpy).toHaveBeenCalledWith(
        expect.stringContaining('error=Test error')
      );
      
      consoleSpy.mockRestore();
    });
  });

  describe('Remote logging', () => {
    it('should send logs to remote endpoint', async () => {
      const fetchMock = fetch as jest.MockedFunction<typeof fetch>;
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
      } as Response);

      const remoteLogger = new Logger({
        level: LogLevel.INFO,
        enableConsole: false,
        enableRemote: true,
        remoteEndpoint: '/api/logs',
        component: 'test-component',
        batchSize: 1, // Force immediate flush
      });

      remoteLogger.info('Test message', 'TestOperation');
      
      // Wait for the async flush to complete
      await new Promise(resolve => setTimeout(resolve, 10));
      
      expect(fetchMock).toHaveBeenCalledWith('/api/logs', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Correlation-ID': expect.any(String),
          'X-User-ID': '',
          'X-Session-ID': expect.any(String),
        },
        body: expect.stringContaining('Test message'),
      });
      
      remoteLogger.destroy();
    });
  });
});

describe('Default logger functions', () => {
  it('should export convenience functions', () => {
    const consoleSpy = jest.spyOn(console, 'info').mockImplementation();
    
    const { info } = require('../utils/logger');
    info('Test message', 'TestOperation', { key: 'value' });
    
    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('Test message')
    );
    
    consoleSpy.mockRestore();
  });
});