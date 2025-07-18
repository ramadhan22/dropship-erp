import { useEffect, useState } from 'react';
import { Logger, LogLevel, startTimer, setCorrelationId, newCorrelationId } from '../utils/logger';

// Example component-specific logger
const reconcileLogger = new Logger({
  level: LogLevel.INFO,
  enableConsole: true,
  enableRemote: true,
  remoteEndpoint: '/api/logs',
  component: 'ReconcileStreamComponent',
  batchSize: 5,
  flushInterval: 3000,
});

interface ReconcileStreamComponentProps {
  shop: string;
  onComplete?: (result: any) => void;
  onError?: (error: Error) => void;
}

interface StreamConfig {
  chunkSize: number;
  maxConcurrency: number;
  memoryThreshold: number;
}

interface StreamResult {
  success: boolean;
  totalProcessed: number;
  totalSuccessful: number;
  totalFailed: number;
  duration: string;
  errorRate: number;
  correlationId: string;
}

const ReconcileStreamComponent: React.FC<ReconcileStreamComponentProps> = ({
  shop,
  onComplete,
  onError,
}) => {
  const [isProcessing, setIsProcessing] = useState(false);
  const [progress, setProgress] = useState<{
    processed: number;
    total: number;
    rate: number;
    eta: string;
  } | null>(null);
  const [config, setConfig] = useState<StreamConfig>({
    chunkSize: 1000,
    maxConcurrency: 5,
    memoryThreshold: 500 * 1024 * 1024, // 500MB
  });

  useEffect(() => {
    // Initialize correlation ID for this component instance
    const correlationId = newCorrelationId();
    setCorrelationId(correlationId);
    
    reconcileLogger.info('Component initialized', 'ComponentMount', {
      shop,
      correlationId,
      config,
    });

    return () => {
      reconcileLogger.info('Component unmounted', 'ComponentUnmount', {
        shop,
        correlationId,
      });
    };
  }, [shop]);

  const handleStreamReconcile = async () => {
    const timer = startTimer('StreamReconciliation');
    const operationId = newCorrelationId();
    setCorrelationId(operationId);

    try {
      setIsProcessing(true);
      
      reconcileLogger.info('Starting stream reconciliation', 'StreamReconcileStart', {
        shop,
        config,
        operationId,
      });

      const response = await fetch('/api/reconcile/stream', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Correlation-ID': operationId,
        },
        body: JSON.stringify({
          shop,
          ...config,
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const result: StreamResult = await response.json();
      
      timer.finish('Stream reconciliation completed successfully', {
        totalProcessed: result.totalProcessed,
        totalSuccessful: result.totalSuccessful,
        totalFailed: result.totalFailed,
        errorRate: result.errorRate,
        duration: result.duration,
      });

      reconcileLogger.info('Stream reconciliation completed', 'StreamReconcileSuccess', {
        shop,
        operationId,
        result: {
          totalProcessed: result.totalProcessed,
          totalSuccessful: result.totalSuccessful,
          totalFailed: result.totalFailed,
          errorRate: result.errorRate,
          duration: result.duration,
        },
      });

      if (onComplete) {
        onComplete(result);
      }
    } catch (err) {
      const errorObj = err as Error;
      
      timer.finishWithError('Stream reconciliation failed', errorObj, {
        shop,
        operationId,
      });

      reconcileLogger.error('Stream reconciliation failed', errorObj, 'StreamReconcileError', {
        shop,
        operationId,
        errorMessage: errorObj.message,
        stack: errorObj.stack,
      });

      if (onError) {
        onError(errorObj);
      }
    } finally {
      setIsProcessing(false);
      setProgress(null);
    }
  };

  const handleConfigChange = (field: keyof StreamConfig, value: number) => {
    setConfig(prev => {
      const newConfig = { ...prev, [field]: value };
      
      reconcileLogger.debug('Configuration changed', 'ConfigUpdate', {
        shop,
        field,
        oldValue: prev[field],
        newValue: value,
        newConfig,
      });
      
      return newConfig;
    });
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow-lg">
      <h2 className="text-2xl font-bold mb-6">Stream Reconciliation</h2>
      
      <div className="mb-6">
        <h3 className="text-lg font-semibold mb-3">Configuration</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Chunk Size
            </label>
            <input
              type="number"
              value={config.chunkSize}
              onChange={(e) => handleConfigChange('chunkSize', parseInt(e.target.value))}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              min="100"
              max="10000"
              step="100"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Max Concurrency
            </label>
            <input
              type="number"
              value={config.maxConcurrency}
              onChange={(e) => handleConfigChange('maxConcurrency', parseInt(e.target.value))}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              min="1"
              max="20"
              step="1"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Memory Threshold (MB)
            </label>
            <input
              type="number"
              value={config.memoryThreshold / (1024 * 1024)}
              onChange={(e) => handleConfigChange('memoryThreshold', parseInt(e.target.value) * 1024 * 1024)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              min="100"
              max="2000"
              step="100"
            />
          </div>
        </div>
      </div>

      {progress && (
        <div className="mb-6 p-4 bg-blue-50 rounded-lg">
          <h3 className="text-lg font-semibold mb-3">Progress</h3>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <div className="text-sm text-gray-600">Processed</div>
              <div className="text-lg font-semibold">{progress.processed.toLocaleString()}</div>
            </div>
            <div>
              <div className="text-sm text-gray-600">Total</div>
              <div className="text-lg font-semibold">{progress.total.toLocaleString()}</div>
            </div>
            <div>
              <div className="text-sm text-gray-600">Rate</div>
              <div className="text-lg font-semibold">{progress.rate.toFixed(1)}/sec</div>
            </div>
            <div>
              <div className="text-sm text-gray-600">ETA</div>
              <div className="text-lg font-semibold">{progress.eta}</div>
            </div>
          </div>
          <div className="mt-3">
            <div className="bg-gray-200 rounded-full h-2">
              <div
                className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                style={{ width: `${(progress.processed / progress.total) * 100}%` }}
              />
            </div>
          </div>
        </div>
      )}

      <div className="flex justify-center">
        <button
          onClick={handleStreamReconcile}
          disabled={isProcessing}
          className={`px-6 py-3 rounded-md font-medium transition-colors ${
            isProcessing
              ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
              : 'bg-blue-500 text-white hover:bg-blue-600'
          }`}
        >
          {isProcessing ? 'Processing...' : 'Start Stream Reconciliation'}
        </button>
      </div>
    </div>
  );
};

export default ReconcileStreamComponent;