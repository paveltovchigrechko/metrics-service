File: server
Type: cpu
Time: 2026-10-09 09:49:44 CEST
Duration: 60.35s, Total samples = 48630ms (80.58%)
Showing nodes accounting for 0, 3.4e-07% of 48630ms total
      flat  flat%   sum%        cum   cum%
         0     0%     0%   -31750ms 65.29%  text/template.(*Template).Execute (inline)
         0     0%     0%   -31750ms 65.29%  text/template.(*Template).execute
    -100ms  0.21%  0.21%   -31750ms 65.29%  text/template.(*state).walk
     -30ms 0.062%  0.27%   -31750ms 65.29%  text/template.(*state).walkIfOrWith
     -10ms 0.021%  0.29%   -31750ms 65.29%  text/template.(*state).walkRange
     -30ms 0.062%  0.35%   -31710ms 65.21%  text/template.(*state).walkRange.func2
         0     0%  0.35% -24402.88ms 50.18%  github.com/paveltovchigrechko/metrics-service/internal/handler.(*AppHandler).MainPage
     -20ms 0.041%  0.39% -21907.23ms 45.05%  github.com/paveltovchigrechko/metrics-service/internal/logger.(*loggingResponseWriter).Write
     -20ms 0.041%  0.43% -21890.79ms 45.01%  net/http.(*response).Write
     -60ms  0.12%  0.56% -21870.79ms 44.97%  net/http.(*response).write
  -46.45ms 0.096%  0.65% -21737.23ms 44.70%  bufio.(*Writer).Write
         0     0%  0.65% -21660.12ms 44.54%  net/http.(*chunkWriter).Write
         0     0%  0.65% -19562.39ms 40.23%  bufio.(*Writer).Flush
         0     0%  0.65% -18704.27ms 38.46%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0%  0.65% -18704.27ms 38.46%  github.com/paveltovchigrechko/metrics-service/internal/middleware.GZIPMiddleware.func1
         0     0%  0.65% -17393.36ms 35.77%  net/http.checkConnErrorWriter.Write
         0     0%  0.65% -17383.36ms 35.75%  net.(*conn).Write
         0     0%  0.65% -17383.36ms 35.75%  net.(*netFD).Write
         0     0%  0.65% -14085.57ms 28.96%  main.run.LoggerMiddleware.func1.1
    3.55ms 0.0073%  0.64% -14085.57ms 28.96%  net/http.HandlerFunc.ServeHTTP
         0     0%  0.64% -14082.01ms 28.96%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0%  0.64% -14082.01ms 28.96%  net/http.serverHandler.ServeHTTP
     -10ms 0.021%  0.66% -12598.10ms 25.91%  internal/poll.(*FD).Write
         0     0%  0.66% -12578.10ms 25.86%  syscall.Write (inline)
         0     0%  0.66% -12578.10ms 25.86%  syscall.write
     -10ms 0.021%  0.68%  9365.88ms 19.26%  net.(*conn).Read
         0     0%  0.68%  9340.33ms 19.21%  internal/poll.(*FD).Read
         0     0%  0.68%  9340.33ms 19.21%  net.(*netFD).Read
         0     0%  0.68%  9340.33ms 19.21%  syscall.Read (inline)
    3.55ms 0.0073%  0.68%  9340.33ms 19.21%  syscall.read
         0     0%  0.68% -6555.81ms 13.48%  net/http.(*conn).serve
         0     0%  0.68%  5645.30ms 11.61%  github.com/paveltovchigrechko/metrics-service/internal/handler.(*AppHandler).UpdatesEndpoint
         0     0%  0.68%  5192.87ms 10.68%  encoding/json.(*Decoder).Decode
         0     0%  0.68%  5178.65ms 10.65%  github.com/paveltovchigrechko/metrics-service/internal/handler.decodeJSONBatch
    3.55ms 0.0073%  0.67%  4788.82ms  9.85%  os.(*File).Write
         0     0%  0.67%  4785.27ms  9.84%  os.(*File).write (inline)
         0     0%  0.67%  4783.04ms  9.84%  go.uber.org/zap/zapcore.(*lockedWriteSyncer).Write
         0     0%  0.67%  4696.76ms  9.66%  net/http.(*connReader).Read
         0     0%  0.67%  4669.12ms  9.60%  net/http.(*connReader).backgroundRead
         0     0%  0.67%  4645.39ms  9.55%  go.uber.org/zap.(*SugaredLogger).Infow
    3.55ms 0.0073%  0.66%  4645.39ms  9.55%  go.uber.org/zap.(*SugaredLogger).log
         0     0%  0.66%  4626.54ms  9.51%  go.uber.org/zap/zapcore.(*CheckedEntry).Write
         0     0%  0.66%  4626.54ms  9.51%  go.uber.org/zap/zapcore.(*ioCore).Write
     -40ms 0.082%  0.74%    -4510ms  9.27%  text/template.(*state).evalPipeline
     -10ms 0.021%  0.77%    -4440ms  9.13%  text/template.(*state).evalCommand
     -30ms 0.062%  0.83%    -4280ms  8.80%  text/template.(*state).evalFunction
         0     0%  0.83%  4184.96ms  8.61%  net/http.(*response).finishRequest
    -150ms  0.31%  1.14%    -4010ms  8.25%  text/template.(*state).evalCall
     -20ms 0.041%  1.18%    -3310ms  6.81%  text/template.(*state).printValue
-3255.53ms  6.69%  7.87% -3255.53ms  6.69%  syscall.rawsyscalln
         0     0%  7.87% -3241.32ms  6.67%  syscall.syscall
         0     0%  7.87% -3241.32ms  6.67%  syscall.syscalln
         0     0%  7.87% -3237.76ms  6.66%  internal/poll.ignoringEINTRIO (inline)
         0     0%  7.87%  3192.87ms  6.57%  bufio.(*Reader).Peek
         0     0%  7.87%  3192.87ms  6.57%  bufio.(*Reader).fill
         0     0%  7.87%  3171.67ms  6.52%  encoding/json.(*decodeState).unmarshal
   39.76ms 0.082%  7.79%  3168.11ms  6.51%  encoding/json.(*decodeState).value
    3.55ms 0.0073%  7.78%  3161.01ms  6.50%  encoding/json.(*decodeState).array
         0     0%  7.78% -3118.67ms  6.41%  fmt.Fprint
  212.12ms  0.44%  7.35%  2917.56ms  6.00%  encoding/json.(*decodeState).object
     -20ms 0.041%  7.39%    -2760ms  5.68%  text/template.safeCall
     -10ms 0.021%  7.41%    -2740ms  5.63%  reflect.Value.Call
    -480ms  0.99%  8.39%    -2730ms  5.61%  reflect.Value.call
 2380.69ms  4.90%  3.50%  2380.69ms  4.90%  runtime.usleep
   70.61ms  0.15%  3.35%  2017.65ms  4.15%  encoding/json.(*Decoder).readValue
   95.97ms   0.2%  3.16%  1844.68ms  3.79%  github.com/paveltovchigrechko/metrics-service/internal/handler.appendMainPage
    8.43ms 0.017%  3.14%  1785.80ms  3.67%  runtime.lock2
         0     0%  3.14%  1785.80ms  3.67%  runtime.lockWithRank (inline)
         0     0%  3.14%  1760.92ms  3.62%  runtime.lock (partial-inline)
    3.55ms 0.0073%  3.13%  1613.65ms  3.32%  encoding/json.(*Decoder).refill
         0     0%  3.13%  1608.14ms  3.31%  runtime.osyield (inline)
         0     0%  3.13%     1571ms  3.23%  io.(*multiReader).Read
         0     0%  3.13%  1503.89ms  3.09%  bufio.(*Reader).Read
         0     0%  3.13%  1503.89ms  3.09%  io.(*LimitedReader).Read
         0     0%  3.13%     1491ms  3.07%  net/http.(*body).Read
         0     0%  3.13%     1491ms  3.07%  net/http.(*body).readLocked
    4.22ms 0.0087%  3.12%  1319.95ms  2.71%  runtime.schedule
    3.55ms 0.0073%  3.12%  1276.87ms  2.63%  runtime.goschedImpl
         0     0%  3.12%  1268.20ms  2.61%  runtime.morestack
         0     0%  3.12%  1267.29ms  2.61%  runtime.newstack
    7.11ms 0.015%  3.10%  1264.22ms  2.60%  runtime.findRunnable
         0     0%  3.10%  1261.33ms  2.59%  runtime.gopreempt_m (inline)
   33.07ms 0.068%  3.03%  1161.19ms  2.39%  encoding/json.(*decodeState).literalStore
   46.21ms 0.095%  2.94%  1080.51ms  2.22%  github.com/paveltovchigrechko/metrics-service/internal/handler.appendValue
     -50ms   0.1%  3.04%    -1070ms  2.20%  text/template.(*state).evalArg
   46.21ms 0.095%  2.95%   951.65ms  1.96%  strconv.AppendFloat (inline)
         0     0%  2.95%   905.45ms  1.86%  internal/strconv.AppendFloat (inline)
   85.97ms  0.18%  2.77%   905.45ms  1.86%  internal/strconv.genericFtoa
     -10ms 0.021%  2.79%   814.49ms  1.67%  runtime.gcstopm
  144.58ms   0.3%  2.49%   778.52ms  1.60%  encoding/json.indirect
   -2.89ms 0.0059%  2.50%   769.66ms  1.58%  runtime.runqgrab
         0     0%  2.50%   769.66ms  1.58%  runtime.runqsteal
 -131.57ms  0.27%  2.77%  -748.73ms  1.54%  runtime.mallocgcSmallScanNoHeader
     -10ms 0.021%  2.79%   746.77ms  1.54%  runtime.stealWork
         0     0%  2.79%   744.32ms  1.53%  runtime.mcall
   60.85ms  0.13%  2.66%   722.13ms  1.48%  github.com/paveltovchigrechko/metrics-service/internal/repository.(*MemStorage).GetAllMetrics
  600.68ms  1.24%  1.43%   661.10ms  1.36%  github.com/paveltovchigrechko/metrics-service/internal/handler.appendEscaped (inline)
    3.55ms 0.0073%  1.42%   637.93ms  1.31%  runtime.park_m
  323.44ms  0.67%  0.76%   572.91ms  1.18%  internal/strconv.fixedFtoa
         0     0%  0.76%   524.80ms  1.08%  runtime.preemptM
  524.80ms  1.08%  0.32%   524.80ms  1.08%  runtime.pthread_kill
         0     0%  0.32%   524.80ms  1.08%  runtime.signalM (inline)
         0     0%  0.32%  -486.02ms     1%  runtime.gcDrainMarkWorkerIdle (inline)
   10.66ms 0.022%  0.34%   467.34ms  0.96%  runtime.systemstack
   14.88ms 0.031%  0.37%   466.64ms  0.96%  github.com/paveltovchigrechko/metrics-service/internal/repository.(*MemStorage).SaveBatch
         0     0%  0.37%     -460ms  0.95%  text/template.(*state).evalFieldNode (inline)
     -40ms 0.082%  0.29%     -450ms  0.93%  text/template.(*state).evalFieldChain
         0     0%  0.29%  -437.18ms   0.9%  runtime.sysUsed (inline)
         0     0%  0.29%  -437.18ms   0.9%  runtime.sysUsedOS (inline)
     -50ms   0.1%  0.19%  -436.93ms   0.9%  runtime.makeslice
 -419.22ms  0.86%  0.67%  -419.22ms  0.86%  runtime.madvise
 -179.15ms  0.37%  1.04%  -415.94ms  0.86%  runtime.mallocgc
     -10ms 0.021%  1.06%  -415.67ms  0.85%  runtime.(*mheap).alloc.func1
  287.66ms  0.59%  0.47%   414.05ms  0.85%  encoding/json.(*decodeState).rescanLiteral
     -60ms  0.12%  0.59%     -410ms  0.84%  text/template.(*state).evalField
   -6.45ms 0.013%  0.61%  -395.67ms  0.81%  runtime.(*mheap).allocSpan
  245.19ms   0.5%   0.1%   387.91ms   0.8%  runtime.wbBufFlush1
         0     0%   0.1%   366.59ms  0.75%  runtime.wbBufFlush
    3.55ms 0.0073% 0.095%   366.59ms  0.75%  runtime.wbBufFlush.func1
         0     0% 0.095%   365.19ms  0.75%  runtime.gcAssistAlloc.func2
         0     0% 0.095%   365.19ms  0.75%  runtime.gcAssistAlloc1
     -50ms   0.1%   0.2%     -360ms  0.74%  reflect.MakeSlice
         0     0%   0.2%   351.82ms  0.72%  runtime.suspendG
         0     0%   0.2%   350.97ms  0.72%  runtime.gcDrainN
  289.89ms   0.6%   0.4%   335.19ms  0.69%  encoding/json.unquoteBytes
    3.55ms 0.0073%  0.41%   334.53ms  0.69%  runtime.wbMove
         0     0%  0.41%  -332.81ms  0.68%  runtime.gcBgMarkWorker
    7.11ms 0.015%  0.42%   324.29ms  0.67%  github.com/paveltovchigrechko/metrics-service/internal/repository.(*MemStorage).saveMetricsLocked (inline)
   -2.89ms 0.0059%  0.41%   319.72ms  0.66%  runtime.markroot
   14.22ms 0.029%  0.44%   309.65ms  0.64%  runtime.mapIterNext
  167.72ms  0.34%  0.79%   305.43ms  0.63%  internal/runtime/maps.(*Iter).Next
    3.55ms 0.0073%   0.8%   302.19ms  0.62%  runtime.markroot.func1
     -20ms 0.041%  0.75%     -290ms   0.6%  fmt.Sprintf
         0     0%  0.75%  -289.99ms   0.6%  runtime.mPark (inline)
         0     0%  0.75%  -289.99ms   0.6%  runtime.notesleep
         0     0%  0.75%   288.72ms  0.59%  runtime.gcDrainMarkWorkerDedicated (inline)
  -26.45ms 0.054%   0.7%  -281.21ms  0.58%  runtime.netpoll
    3.55ms 0.0073%  0.71%  -273.12ms  0.56%  runtime.stopm
    3.55ms 0.0073%  0.71%   272.78ms  0.56%  runtime.gcAssistAlloc
   93.07ms  0.19%  0.91%   272.78ms  0.56%  encoding/json.(*decodeState).scanWhile
   -9.34ms 0.019%  0.89%   263.44ms  0.54%  runtime.deductAssistCredit
  -12.23ms 0.025%  0.86%   260.97ms  0.54%  runtime.bulkBarrierPreWrite
   11.33ms 0.023%  0.89%   256.57ms  0.53%  internal/strconv.formatDigits
 -251.21ms  0.52%  0.37%  -251.21ms  0.52%  runtime.kevent
  213.26ms  0.44%  0.81%   245.25ms   0.5%  internal/strconv.fmtF
     -60ms  0.12%  0.68%     -240ms  0.49%  text/template.findFunction
  234.58ms  0.48%  1.17%   234.58ms  0.48%  internal/strconv.formatBase10
    -100ms  0.21%  0.96%     -230ms  0.47%  reflect.funcLayout
    3.55ms 0.0073%  0.97%   225.67ms  0.46%  runtime.(*wbBuf).get2 (inline)
  139.04ms  0.29%  1.25%   222.12ms  0.46%  reflect.(*rtype).Name
     -40ms 0.082%  1.17%     -220ms  0.45%  reflect.(*rtype).FieldByName
         0     0%  1.17%  -217.30ms  0.45%  runtime.gcBgMarkWorker.func2
     -60ms  0.12%  1.05%  -216.45ms  0.45%  fmt.(*pp).printArg
         0     0%  1.05%   198.80ms  0.41%  github.com/paveltovchigrechko/metrics-service/internal/repository.metricKey (inline)
   14.22ms 0.029%  1.08%   198.80ms  0.41%  runtime.concatstring3
    0.66ms 0.0014%  1.08%  -197.30ms  0.41%  runtime.gcDrain
  100.18ms  0.21%  1.28%   194.58ms   0.4%  runtime.concatstrings
  130.61ms  0.27%  1.55%   194.34ms   0.4%  encoding/json.stateEndValue
 -192.51ms   0.4%  1.16%  -192.51ms   0.4%  runtime.pthread_cond_wait
     -30ms 0.062%  1.10%     -190ms  0.39%  fmt.(*pp).doPrintf
         0     0%  1.10%  -183.13ms  0.38%  go.uber.org/zap/zapcore.(*jsonEncoder).AddString
         0     0%  1.10%  -183.13ms  0.38%  go.uber.org/zap/zapcore.(*jsonEncoder).AppendString
         0     0%  1.10%   181.41ms  0.37%  runtime.preemptone
     -10ms 0.021%  1.08%     -180ms  0.37%  callRet
     -90ms  0.19%  0.89%     -180ms  0.37%  reflect.(*structType).FieldByName
         0     0%  0.89%  -179.58ms  0.37%  go.uber.org/zap/zapcore.(*jsonEncoder).safeAddString (inline)
  -93.79ms  0.19%   0.7%  -179.58ms  0.37%  go.uber.org/zap/zapcore.safeAppendStringLike[go.shape.string]
     -20ms 0.041%  0.66%   177.42ms  0.36%  runtime.newobject
    3.55ms 0.0073%  0.66%  -174.70ms  0.36%  go.uber.org/zap/zapcore.consoleEncoder.writeContext
    -110ms  0.23%  0.44%     -170ms  0.35%  runtime.reflectcallmove
    3.55ms 0.0073%  0.44%  -165.36ms  0.34%  go.uber.org/zap/zapcore.Field.AddTo
         0     0%  0.44%  -164.46ms  0.34%  runtime.pollWork
 -151.57ms  0.31%  0.13%  -163.79ms  0.34%  runtime.(*mspan).writeHeapBitsSmall
         0     0%  0.13%  -163.79ms  0.34%  runtime.heapSetTypeNoHeader (inline)
    3.55ms 0.0073%  0.14%   163.26ms  0.34%  strconv.ParseFloat
    3.55ms 0.0073%  0.15%  -161.81ms  0.33%  go.uber.org/zap/zapcore.addFields (inline)
         0     0%  0.15%     -160ms  0.33%  reflect.unsafe_NewArray
    7.11ms 0.015%  0.16%   159.70ms  0.33%  internal/strconv.ParseFloat
         0     0%  0.16%   157.07ms  0.32%  runtime.preemptall
         0     0%  0.16%  -156.54ms  0.32%  runtime.semasleep
    3.55ms 0.0073%  0.17%  -156.50ms  0.32%  go.uber.org/zap/zapcore.consoleEncoder.EncodeEntry
         0     0%  0.17%   154.82ms  0.32%  net/http.(*conn).readRequest
         0     0%  0.17%  -152.89ms  0.31%  runtime.newarray
    7.11ms 0.015%  0.18%   152.60ms  0.31%  internal/strconv.parseFloatPrefix
     -10ms 0.021%  0.16%     -150ms  0.31%  text/template.printableValue
         0     0%  0.16%      147ms   0.3%  runtime.runSafePointFn
     -40ms 0.082% 0.081%  -146.45ms   0.3%  net/http.(*conn).hijacked
   17.77ms 0.037%  0.12%   145.49ms   0.3%  internal/strconv.atof64
     -10ms 0.021% 0.097%  -144.94ms   0.3%  reflect.unsafe_New
   73.07ms  0.15%  0.25%   142.60ms  0.29%  internal/runtime/maps.(*Map).getWithoutKeySmallFastStr
    8.43ms 0.017%  0.26%   140.79ms  0.29%  runtime.slicebytetostring
    3.55ms 0.0073%  0.27%   140.20ms  0.29%  runtime.forEachPInternal
     -20ms 0.041%  0.23%     -140ms  0.29%  reflect.intToReg
         0     0%  0.23%  -136.45ms  0.28%  runtime.convT
         0     0%  0.23%   133.50ms  0.27%  net/http.readRequest
   68.86ms  0.14%  0.37%   131.27ms  0.27%  encoding/json.stateBeginValue
     -20ms 0.041%  0.33%     -130ms  0.27%  text/template.(*state).evalEmptyInterface
     -10ms 0.021%  0.31%   128.38ms  0.26%  runtime.growslice
  143.26ms  0.29%  0.61%   128.14ms  0.26%  runtime.findObject
   60.85ms  0.13%  0.73%   125.25ms  0.26%  runtime.mallocgcTiny
   91.51ms  0.19%  0.92%   123.50ms  0.25%  encoding/json.stateInString
     -20ms 0.041%  0.88%     -120ms  0.25%  text/template.(*state).at (inline)
     -40ms 0.082%   0.8%     -120ms  0.25%  text/template.isMissing
         0     0%   0.8%   118.80ms  0.24%  runtime.stopTheWorldWithSema
         0     0%   0.8%   116.81ms  0.24%  runtime.ready
         0     0%   0.8%   116.39ms  0.24%  reflect.Value.Grow
    3.55ms 0.0073%   0.8%   116.39ms  0.24%  reflect.Value.grow
         0     0%   0.8%   115.06ms  0.24%  runtime.unlock2Wake
  112.84ms  0.23%  1.03%   112.84ms  0.23%  internal/strconv.readFloat
     -20ms 0.041%  0.99%     -110ms  0.23%  runtime.typedmemmove
     -40ms 0.082%  0.91%     -110ms  0.23%  text/template.(*state).validateType
         0     0%  0.91%   109.28ms  0.22%  reflect.growslice
         0     0%  0.91%   108.62ms  0.22%  runtime.unlock (partial-inline)
     -10ms 0.021%  0.89%   108.62ms  0.22%  runtime.unlock2
         0     0%  0.89%   108.62ms  0.22%  runtime.unlockWithRank (inline)
  105.73ms  0.22%  1.11%   105.73ms  0.22%  runtime.asyncPreempt
    3.55ms 0.0073%  1.12%   104.47ms  0.21%  runtime.wakep
  -91.73ms  0.19%  0.93%  -103.78ms  0.21%  runtime.tryDeferToSpanScan
         0     0%  0.93%   101.27ms  0.21%  runtime.readyWithTime.goready.func1
     -20ms 0.041%  0.89%     -100ms  0.21%  fmt.(*fmt).fmtFloat
         0     0%  0.89%     -100ms  0.21%  fmt.(*pp).fmtFloat
   -1.81ms 0.0037%  0.88%    99.35ms   0.2%  gcWriteBarrier
  -98.01ms   0.2%  0.68%   -98.01ms   0.2%  runtime.memclrNoHeapPointers
         0     0%  0.68%    97.12ms   0.2%  runtime.semawakeup
   74.40ms  0.15%  0.83%    97.05ms   0.2%  runtime.mapaccess2_faststr
         0     0%  0.83%    94.40ms  0.19%  runtime.procyield (inline)
   94.40ms  0.19%  1.03%    94.40ms  0.19%  runtime.procyieldAsm
   93.57ms  0.19%  1.22%    93.57ms  0.19%  runtime.pthread_cond_signal
     -20ms 0.041%  1.18%   -92.47ms  0.19%  runtime.(*mcache).nextFree
     -10ms 0.021%  1.16%    91.93ms  0.19%  runtime.gcStart.func2
  -24.46ms  0.05%  1.11%    91.69ms  0.19%  runtime.mapaccess1_faststr
     -90ms  0.19%  0.92%      -90ms  0.19%  internal/abi.(*RegArgs).IntRegArgAddr (inline)
    3.55ms 0.0073%  0.93%    89.94ms  0.18%  reflect.New
         0     0%  0.93%    89.52ms  0.18%  runtime.exitsyscallNoP
     -10ms 0.021%  0.91%   -88.67ms  0.18%  runtime.convTstring
   21.33ms 0.044%  0.95%    87.29ms  0.18%  encoding/json.cachedTypeFields
         0     0%  0.95%    87.12ms  0.18%  runtime.gcMarkDone.forEachP.func5
  -83.13ms  0.17%  0.78%   -83.13ms  0.17%  runtime.getMCache (inline)
   46.21ms 0.095%  0.88%    83.07ms  0.17%  encoding/json.stateBeginString
   17.77ms 0.037%  0.91%    83.07ms  0.17%  reflect.(*rtype).String
   -6.45ms 0.013%   0.9%   -82.89ms  0.17%  fmt.(*pp).doPrint
     -30ms 0.062%  0.84%      -80ms  0.16%  fmt.newPrinter
         0     0%  0.84%      -80ms  0.16%  io.ReadAll
  -82.23ms  0.17%  0.67%   -78.67ms  0.16%  runtime.ifaceeq
   19.10ms 0.039%  0.71%    78.62ms  0.16%  reflect.Value.Addr
    3.55ms 0.0073%  0.72%    76.63ms  0.16%  runtime.rawstringtmp
  -96.45ms   0.2%  0.52%   -74.46ms  0.15%  internal/sync.(*Mutex).Lock (inline)
         0     0%  0.52%   -74.46ms  0.15%  sync.(*Mutex).Lock (inline)
     -20ms 0.041%  0.48%   -73.13ms  0.15%  runtime.(*mcache).refill
         0     0%  0.48%    73.07ms  0.15%  runtime.rawstring (inline)
    3.55ms 0.0073%  0.48%   -71.57ms  0.15%  runtime.(*mcentral).grow
         0     0%  0.48%    70.85ms  0.15%  reflect.ptrTo (inline)
         0     0%  0.48%      -70ms  0.14%  reflect.typedmemmove
  -22.89ms 0.047%  0.44%   -69.34ms  0.14%  go.uber.org/zap/buffer.(*Buffer).AppendString
         0     0%  0.44%    64.64ms  0.13%  runtime.mallocgcLarge
  -22.23ms 0.046%  0.39%   -64.03ms  0.13%  internal/sync.(*HashTrieMap[go.shape.interface {},go.shape.interface {}]).Load
         0     0%  0.39%   -64.03ms  0.13%  sync.(*Map).Load (inline)
    3.55ms 0.0073%   0.4%   -63.79ms  0.13%  runtime.(*mcentral).cacheSpan
     -10ms 0.021%  0.38%   -63.37ms  0.13%  runtime.scanframeworker
     -30ms 0.062%  0.32%      -60ms  0.12%  reflect.(*structType).Field
         0     0%  0.32%      -60ms  0.12%  reflect.Value.Interface (inline)
         0     0%  0.32%      -60ms  0.12%  reflect.valueInterface
     -40ms 0.082%  0.23%      -60ms  0.12%  text/template.eq
   26.21ms 0.054%  0.29%    59.52ms  0.12%  runtime.mallocgcSmallNoscan
  -59.34ms  0.12%  0.17%   -59.34ms  0.12%  runtime.divRoundUp (inline)
         0     0%  0.17%    58.19ms  0.12%  reflect.(*rtype).nameOff (inline)
   21.33ms 0.044%  0.21%    58.19ms  0.12%  reflect.resolveNameOff
   -8.01ms 0.016%  0.19%   -54.46ms  0.11%  internal/sync.(*entry[go.shape.interface {},go.shape.interface {}]).lookup (inline)
    3.55ms 0.0073%   0.2%    53.98ms  0.11%  runtime.bulkBarrierPreWriteSrcOnly
    4.88ms  0.01%  0.21%    53.50ms  0.11%  runtime.(*sweepLocked).sweep
         0     0%  0.21%   -53.19ms  0.11%  runtime.scanstack
         0     0%  0.21%    53.07ms  0.11%  runtime.gcMarkTermination.forEachP.func7
         0     0%  0.21%      -50ms   0.1%  reflect.(*rtype).Implements
     -40ms 0.082%  0.13%      -50ms   0.1%  reflect.implements
     -40ms 0.082% 0.047%      -50ms   0.1%  reflect.memmove
         0     0% 0.047%      -50ms   0.1%  reflect.packEface (inline)
         0     0% 0.047%      -50ms   0.1%  reflect.packEfaceData
     -30ms 0.062% 0.015%      -50ms   0.1%  runtime.efaceeq
         0     0% 0.015%      -50ms   0.1%  text/template.(*Template).Parse
         0     0% 0.015%   -49.34ms   0.1%  reflect.(*rtype).Elem (partial-inline)
  -42.89ms 0.088%   0.1%   -49.34ms   0.1%  reflect.elem
  -39.39ms 0.081%  0.18%   -48.79ms   0.1%  runtime.scanObjectsSmall
   48.19ms 0.099% 0.085%    48.19ms 0.099%  aeshashbody
  -29.34ms  0.06%  0.15%   -48.01ms 0.099%  runtime.(*spanInlineMarkBits).tryAcquire
   42.65ms 0.088% 0.058%    46.87ms 0.096%  encoding/json.(*decodeState).scanNext
         0     0% 0.058%    46.21ms 0.095%  runtime.(*mheap).allocManual
    3.55ms 0.0073%  0.05%    46.21ms 0.095%  strconv.AppendInt (inline)
         0     0%  0.05%    45.30ms 0.093%  runtime.preemptPark
   -6.45ms 0.013% 0.064%   -42.89ms 0.088%  fmt.(*pp).free
     -30ms 0.062%  0.13%   -42.89ms 0.088%  internal/abi.Name.Name
         0     0%  0.13%    42.65ms 0.088%  internal/strconv.AppendInt
   21.33ms 0.044% 0.082%    42.65ms 0.088%  internal/strconv.AppendUint
   14.22ms 0.029% 0.052%    42.65ms 0.088%  runtime.(*wbBuf).get1 (inline)
         0     0% 0.052%    42.65ms 0.088%  runtime.getempty
         0     0% 0.052%    42.65ms 0.088%  runtime.getempty.func1
         0     0% 0.052%    42.41ms 0.087%  runtime.(*mheap).freeSpan (inline)
   18.43ms 0.038% 0.014%    40.85ms 0.084%  reflect.(*rtype).ptrTo
   11.33ms 0.023% 0.0089%    40.42ms 0.083%  encoding/json.stateBeginStringOrEmpty
     -20ms 0.041% 0.032%      -40ms 0.082%  fmt.(*buffer).write (inline)
         0     0% 0.032%      -40ms 0.082%  fmt.(*fmt).pad
     -40ms 0.082%  0.11%      -40ms 0.082%  internal/abi.(*IntArgRegBitmap).Get (inline)
     -40ms 0.082%   0.2%      -40ms 0.082%  reflect.(*abiSeq).stepsForValue (inline)
         0     0%   0.2%      -40ms 0.082%  runtime.(*mspan).initHeapBits
         0     0%   0.2%      -40ms 0.082%  text/template/parse.(*Tree).Parse
         0     0%   0.2%      -40ms 0.082%  text/template/parse.Parse
         0     0%   0.2%    39.10ms  0.08%  github.com/paveltovchigrechko/metrics-service/internal/handler.(*AppHandler).ValueEndpoint
         0     0%   0.2%    38.19ms 0.079%  runtime.(*mcache).prepareForSweep
   -2.89ms 0.0059%   0.2%    38.19ms 0.079%  runtime.(*mcache).releaseAll
   16.21ms 0.033%  0.17%    38.19ms 0.079%  runtime.resolveTypeOff
         0     0%  0.17%    38.19ms 0.079%  runtime.rtype.typeOff (inline)
   36.87ms 0.076% 0.094%    36.87ms 0.076%  runtime.resolveNameOff
    3.55ms 0.0073% 0.086%    36.21ms 0.074%  runtime.(*mheap).reclaim
         0     0% 0.086%    35.97ms 0.074%  runtime.notetsleep
         0     0% 0.086%    35.97ms 0.074%  runtime.notetsleep_internal
   35.97ms 0.074% 0.012%    35.97ms 0.074%  runtime.pthread_cond_timedwait_relative_np
         0     0% 0.012%   -35.78ms 0.074%  runtime.copystack
         0     0% 0.012%    35.54ms 0.073%  runtime.(*mcache).allocLarge
  -19.34ms  0.04% 0.052%   -35.12ms 0.072%  reflect.Value.assignTo
     -10ms 0.021% 0.073%    34.40ms 0.071%  runtime.sweepone
         0     0% 0.073%    33.98ms  0.07%  runtime.(*mcentral).uncacheSpan
   19.28ms  0.04% 0.033%    33.50ms 0.069%  runtime.typePointers.next
   -2.89ms 0.0059% 0.039%   -33.13ms 0.068%  runtime.(*stkframe).getStackMap
         0     0% 0.039%    33.07ms 0.068%  runtime.(*sweepLocked).sweep.(*mheap).freeSpan.func2
         0     0% 0.039%   -32.89ms 0.068%  sync.(*Pool).Put
     -30ms 0.062%   0.1%   -32.89ms 0.068%  sync.(*RWMutex).RLock (inline)
  -25.78ms 0.053%  0.15%   -32.23ms 0.066%  runtime.greyobject
   21.33ms 0.044%  0.11%    31.99ms 0.066%  github.com/paveltovchigrechko/metrics-service/internal/model.ValidateMetrics
         0     0%  0.11%    31.99ms 0.066%  runtime.(*stackScanState).putPtr
         0     0%  0.11%    31.99ms 0.066%  slices.Grow[go.shape.[]uint8,go.shape.uint8] (inline)
  -15.12ms 0.031%  0.14%   -30.90ms 0.064%  runtime.step
    3.55ms 0.0073%  0.13%   -30.78ms 0.063%  runtime.(*gcControllerState).enlistWorker
     -30ms 0.062%   0.2%      -30ms 0.062%  reflect.(*rtype).AssignableTo
         0     0%   0.2%      -30ms 0.062%  reflect.PointerTo (inline)
     -10ms 0.021%  0.22%      -30ms 0.062%  reflect.Value.FieldByIndexErr
     -20ms 0.041%  0.26%      -30ms 0.062%  reflect.Value.MethodByName
     -10ms 0.021%  0.28%      -30ms 0.062%  reflect.intFromReg
         0     0%  0.28%      -30ms 0.062%  text/template/parse.(*Tree).action
         0     0%  0.28%      -30ms 0.062%  text/template/parse.(*Tree).command
         0     0%  0.28%      -30ms 0.062%  text/template/parse.(*Tree).ifControl
         0     0%  0.28%      -30ms 0.062%  text/template/parse.(*Tree).itemList
         0     0%  0.28%      -30ms 0.062%  text/template/parse.(*Tree).parse
         0     0%  0.28%      -30ms 0.062%  text/template/parse.(*Tree).parseControl
         0     0%  0.28%      -30ms 0.062%  text/template/parse.(*Tree).pipeline
         0     0%  0.28%      -30ms 0.062%  text/template/parse.(*Tree).rangeControl
         0     0%  0.28%      -30ms 0.062%  text/template/parse.(*Tree).textOrAction
   36.21ms 0.074%   0.2%    29.76ms 0.061%  reflect.Value.Elem
         0     0%   0.2%    29.10ms  0.06%  runtime.(*mheap).reclaimChunk
         0     0%   0.2%    29.10ms  0.06%  runtime.injectglist
   28.43ms 0.058%  0.14%    28.43ms 0.058%  encoding/json.(*decodeState).readIndex (inline)
   28.43ms 0.058% 0.086%    28.43ms 0.058%  encoding/json.isSpace (inline)
         0     0% 0.086%    28.43ms 0.058%  runtime.gcFlushBgCredit
         0     0% 0.086%    28.43ms 0.058%  sync.(*Pool).pinSlow
  -56.45ms  0.12%   0.2%   -28.01ms 0.058%  sync.(*Pool).pin
   27.53ms 0.057%  0.15%    27.53ms 0.057%  runtime.acquirem (inline)
   23.98ms 0.049% 0.096%    27.53ms 0.057%  unicode/utf8.DecodeRune (inline)
         0     0% 0.096%   -26.45ms 0.054%  fmt.Fprintf
         0     0% 0.096%   -26.45ms 0.054%  runtime.convT64
   26.21ms 0.054% 0.042%    26.21ms 0.054%  runtime.arenaIndex (inline)
   -2.89ms 0.0059% 0.048%   -25.78ms 0.053%  runtime.(*mheap).initSpan
   -2.89ms 0.0059% 0.054%    25.54ms 0.053%  encoding/json.state1
   17.77ms 0.037% 0.018%    25.54ms 0.053%  encoding/json.stateDot0
   14.22ms 0.029% 0.012%    25.54ms 0.053%  reflect.Value.SetString
         0     0% 0.012%    25.54ms 0.053%  runtime.injectglist.func1
         0     0% 0.012%    25.54ms 0.053%  runtime.semacquire1
   -9.34ms 0.019% 0.0077%   -25.12ms 0.052%  runtime.newMarkBits
  -25.12ms 0.052% 0.059%   -25.12ms 0.052%  runtime.releasem (inline)
         0     0% 0.059%    25.06ms 0.052%  runtime.(*pageAlloc).scavenge.func1
         0     0% 0.059%    25.06ms 0.052%  runtime.(*pageAlloc).scavengeOne
         0     0% 0.059%    24.88ms 0.051%  net/http.Header.Set (inline)
         0     0% 0.059%    24.88ms 0.051%  net/textproto.MIMEHeader.Set (inline)
    7.11ms 0.015% 0.045%    24.88ms 0.051%  runtime.(*mheap).nextSpanForSweep
   21.33ms 0.044% 0.00084%    24.88ms 0.051%  runtime.(*spanSet).pop
         0     0% 0.00084%    24.88ms 0.051%  runtime.gcMarkDone.func1
   -1.57ms 0.0032% 0.0041%    24.64ms 0.051%  runtime.heapArenaOf (inline)
   -6.45ms 0.013% 0.017%   -24.03ms 0.049%  runtime.(*unwinder).next
         0     0% 0.017%    23.31ms 0.048%  runtime.acquirep
         0     0% 0.017%    23.31ms 0.048%  runtime.acquirepNoTrace
   -6.45ms 0.013% 0.031%   -22.89ms 0.047%  go.uber.org/zap/buffer.(*Buffer).Write (partial-inline)
         0     0% 0.031%    22.65ms 0.047%  runtime.gcMarkTermination.func3
         0     0% 0.031%    22.41ms 0.046%  reflect.(*rtype).typeOff (inline)
   -5.78ms 0.012% 0.042%    22.41ms 0.046%  reflect.resolveTypeOff
         0     0% 0.042%    21.99ms 0.045%  internal/sync.(*Mutex).lockSlow
   21.99ms 0.045% 0.0028%    21.99ms 0.045%  reflect.Value.Len (inline)
   21.99ms 0.045% 0.048%    21.99ms 0.045%  runtime.extractHeapBitsSmall
   21.99ms 0.045% 0.093%    21.99ms 0.045%  runtime.mapaccess1_fast32
   -6.45ms 0.013%  0.08%   -21.92ms 0.045%  runtime.scanSpan
   21.33ms 0.044%  0.12%    21.33ms 0.044%  internal/strconv.umul192 (inline)
    3.55ms 0.0073%  0.13%    21.33ms 0.044%  reflect.(*rtype).NumMethod
   21.33ms 0.044%  0.17%    21.33ms 0.044%  reflect.Value.IsNil (inline)
         0     0%  0.17%    21.33ms 0.044%  runtime.(*sweepLocked).sweep.(*mheap).freeSpan.func3
    7.11ms 0.015%  0.19%    21.33ms 0.044%  runtime.(*unwinder).initAt
         0     0%  0.19%    21.33ms 0.044%  runtime.gcParkAssist
         0     0%  0.19%    21.33ms 0.044%  sync.(*RWMutex).Lock
    3.55ms 0.0073%   0.2%   -20.24ms 0.042%  runtime.pcdatavalue
     -10ms 0.021%  0.18%      -20ms 0.041%  fmt.(*pp).fmtInteger
     -20ms 0.041%  0.14%      -20ms 0.041%  internal/abi.(*FuncType).InSlice (inline)
     -10ms 0.021%  0.11%      -20ms 0.041%  reflect.(*rtype).In
     -10ms 0.021% 0.094%      -20ms 0.041%  reflect.(*rtype).IsVariadic
         0     0% 0.094%      -20ms 0.041%  runtime.(*moduledata).funcName
     -10ms 0.021% 0.073%      -20ms 0.041%  runtime.findnull
         0     0% 0.073%      -20ms 0.041%  runtime.funcname (inline)
         0     0% 0.073%      -20ms 0.041%  runtime.gcStart
         0     0% 0.073%      -20ms 0.041%  runtime.gostringnocopy (inline)
         0     0% 0.073%      -20ms 0.041%  runtime.isSystemGoroutine
     -20ms 0.041% 0.032%      -20ms 0.041%  runtime.memequal128
         0     0% 0.032%      -20ms 0.041%  runtime.newproc
     -10ms 0.021% 0.012%      -20ms 0.041%  runtime.typedslicecopy
     -20ms 0.041% 0.029%      -20ms 0.041%  runtime.unspillArgs
         0     0% 0.029%      -20ms 0.041%  text/template/parse.(*Tree).operand
         0     0% 0.029%      -20ms 0.041%  text/template/parse.(*Tree).term
   -8.67ms 0.018% 0.047%    19.76ms 0.041%  runtime.casgstatus
         0     0% 0.047%    19.10ms 0.039%  go.uber.org/zap/zapcore.ISO8601TimeEncoder
    3.55ms 0.0073%  0.04%    19.10ms 0.039%  go.uber.org/zap/zapcore.encodeTimeLayout
         0     0%  0.04%   -18.91ms 0.039%  runtime.newproc.func1
  -18.67ms 0.038% 0.078%   -18.67ms 0.038%  runtime.(*spanScanOwnership).or (inline)
         0     0% 0.078%   -18.67ms 0.038%  runtime.gcStart.func4
  -18.67ms 0.038%  0.12%   -18.67ms 0.038%  runtime.spanOf (inline)
   21.33ms 0.044% 0.073%    18.43ms 0.038%  encoding/json.(*scanner).popParseState (inline)
         0     0% 0.073%    18.43ms 0.038%  internal/sync.runtime_SemacquireMutex
         0     0% 0.073%    18.43ms 0.038%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0% 0.073%    18.43ms 0.038%  net/textproto.readMIMEHeader
   18.43ms 0.038% 0.035%    18.43ms 0.038%  runtime.typePointers.nextFast (inline)
         0     0% 0.035%    17.95ms 0.037%  runtime.sysUnused (inline)
         0     0% 0.035%    17.95ms 0.037%  runtime.sysUnusedOS (inline)
         0     0% 0.035%   -17.94ms 0.037%  runtime.notewakeup
         0     0% 0.035%    17.77ms 0.037%  net/http.Header.WriteSubset (inline)
         0     0% 0.035%    17.77ms 0.037%  net/http.Header.sortedKeyValues
         0     0% 0.035%    17.77ms 0.037%  net/http.Header.writeSubset
   17.77ms 0.037% 0.0017%    17.77ms 0.037%  runtime.(*sweepLocker).tryAcquire
         0     0% 0.0017%    17.77ms 0.037%  runtime.GOMAXPROCS
         0     0% 0.0017%    17.77ms 0.037%  runtime.execute
    7.11ms 0.015% 0.016%    17.77ms 0.037%  strconv.ParseInt
   21.75ms 0.045% 0.061%   -17.16ms 0.035%  runtime.pcvalue
         0     0% 0.061%    16.87ms 0.035%  runtime.gcMarkDone.func2
         0     0% 0.061%   -16.45ms 0.034%  fmt.(*fmt).fmtS
     -20ms 0.041%  0.02%   -16.45ms 0.034%  fmt.(*fmt).padString
         0     0%  0.02%   -16.45ms 0.034%  fmt.(*pp).fmtString
     -10ms 0.021% 0.0007%   -16.45ms 0.034%  go.uber.org/zap/zapcore.ShortCallerEncoder
         0     0% 0.0007%   -16.45ms 0.034%  runtime.goyield_m
         0     0% 0.0007%   -16.45ms 0.034%  runtime.newproc1
         0     0% 0.0007%   -16.45ms 0.034%  runtime.shrinkstack
     -10ms 0.021% 0.021%   -16.45ms 0.034%  sync.(*poolChain).popTail
    7.11ms 0.015% 0.0066%    16.21ms 0.033%  net/http.(*chunkWriter).writeHeader
         0     0% 0.0066%    16.21ms 0.033%  runtime.markrootBlock
         0     0% 0.0066%    16.21ms 0.033%  runtime.nanotime (inline)
   22.65ms 0.047%  0.04%    16.21ms 0.033%  runtime.nanotime1
         0     0%  0.04%    15.97ms 0.033%  runtime.resetspinning
  -15.78ms 0.032% 0.0075%   -15.78ms 0.032%  runtime.readvarint (inline)
    3.55ms 0.0073% 0.015%    15.54ms 0.032%  runtime.entersyscallHandleGCWait
         0     0% 0.015%    15.54ms 0.032%  runtime.gosched_m
         0     0% 0.015%    15.54ms 0.032%  runtime.reentersyscall.func4
   14.88ms 0.031% 0.045%    14.88ms 0.031%  reflect.Value.Type (inline)
         0     0% 0.045%    14.88ms 0.031%  runtime.(*timers).check
         0     0% 0.045%    14.22ms 0.029%  github.com/paveltovchigrechko/metrics-service/internal/handler.decodeJSONMetrics
         0     0% 0.045%    14.22ms 0.029%  github.com/paveltovchigrechko/metrics-service/internal/handler.init.func1
         0     0% 0.045%    14.22ms 0.029%  github.com/paveltovchigrechko/metrics-service/internal/logger.peekBody
   14.22ms 0.029% 0.075%    14.22ms 0.029%  internal/runtime/atomic.(*Uint64).Add (inline)
         0     0% 0.075%    14.22ms 0.029%  net/textproto.(*Reader).ReadLine (inline)
   10.66ms 0.022% 0.097%    14.22ms 0.029%  reflect.(*rtype).exportedMethods
         0     0% 0.097%    14.22ms 0.029%  reflect.Value.Kind (inline)
   14.22ms 0.029%  0.13%    14.22ms 0.029%  runtime.(*timers).wakeTime (inline)
         0     0%  0.13%    14.22ms 0.029%  runtime.(*unwinder).init (inline)
         0     0%  0.13%    14.22ms 0.029%  runtime.entersyscall
    7.11ms 0.015%  0.14%    14.22ms 0.029%  runtime.semrelease1
   14.22ms 0.029%  0.17%    14.22ms 0.029%  runtime.walltime
         0     0%  0.17%    14.22ms 0.029%  time.Now
         0     0%  0.17%    14.22ms 0.029%  time.now
         0     0%  0.17%    14.22ms 0.029%  time.runtimeNow
   43.31ms 0.089%  0.26%    13.74ms 0.028%  runtime.scanblock
         0     0%  0.26%    13.31ms 0.027%  go.uber.org/zap.(*Logger).Check (inline)
     -10ms 0.021%  0.24%    13.31ms 0.027%  go.uber.org/zap.(*Logger).check
   -2.89ms 0.0059%  0.23%   -12.89ms 0.027%  go.uber.org/zap/buffer.(*Buffer).AppendByte (inline)
   -6.45ms 0.013%  0.22%   -12.89ms 0.027%  go.uber.org/zap/buffer.Pool.Get
  -12.89ms 0.027%  0.19%   -12.89ms 0.027%  internal/abi.Name.ReadVarint (inline)
         0     0%  0.19%   -12.89ms 0.027%  net/http.(*connReader).startBackgroundRead
  -12.89ms 0.027%  0.17%   -12.89ms 0.027%  reflect.directlyAssignable
  -12.89ms 0.027%  0.14%   -12.89ms 0.027%  runtime.(*gcBitsArena).tryAlloc (inline)
     -20ms 0.041% 0.098%   -12.89ms 0.027%  runtime.(*stackScanState).addObject
    3.55ms 0.0073%  0.11%   -12.89ms 0.027%  runtime.adjustframe
    3.55ms 0.0073%  0.11%   -12.89ms 0.027%  sync.(*Pool).getSlow
         0     0%  0.11%   -12.23ms 0.025%  time.Time.AppendFormat
  -12.89ms 0.027% 0.086%   -12.23ms 0.025%  time.Time.appendFormat
         0     0% 0.086%    11.99ms 0.025%  go.uber.org/zap/internal/stacktrace.Capture
         0     0% 0.086%    11.99ms 0.025%  runtime.Callers (inline)
   24.88ms 0.051%  0.14%    11.99ms 0.025%  runtime.gcmarknewobject
         0     0%  0.14%    11.99ms 0.025%  runtime.mallocgcSmallScanHeader
   11.99ms 0.025%  0.16%    11.99ms 0.025%  runtime.spanSetScans
         0     0%  0.16%   -11.57ms 0.024%  runtime.mProf_Malloc
         0     0%  0.16%   -11.57ms 0.024%  runtime.profilealloc
   14.22ms 0.029%  0.19%    11.33ms 0.023%  runtime.typehash
         0     0%  0.19%    10.66ms 0.022%  encoding/json.Marshal
    7.11ms 0.015%  0.21%    10.66ms 0.022%  encoding/json.stateDot
         0     0%  0.21%    10.66ms 0.022%  github.com/paveltovchigrechko/metrics-service/internal/logger.(*loggingResponseWriter).WriteHeader (partial-inline)
         0     0%  0.21%    10.66ms 0.022%  github.com/paveltovchigrechko/metrics-service/internal/model.ValidateName (inline)
         0     0%  0.21%    10.66ms 0.022%  go.uber.org/zap/zapcore.(*sliceArrayEncoder).AppendString
         0     0%  0.21%    10.66ms 0.022%  internal/poll.(*FD).SetReadDeadline (inline)
         0     0%  0.21%    10.66ms 0.022%  internal/poll.setDeadlineImpl
   10.66ms 0.022%  0.23%    10.66ms 0.022%  internal/runtime/atomic.(*Int32).Load (inline)
   10.66ms 0.022%  0.25%    10.66ms 0.022%  internal/runtime/maps.(*groupReference).key (inline)
         0     0%  0.25%    10.66ms 0.022%  internal/runtime/maps.NewEmptyMap (inline)
         0     0%  0.25%    10.66ms 0.022%  internal/strconv.ParseInt
   10.66ms 0.022%  0.27%    10.66ms 0.022%  internal/strconv.ParseUint
         0     0%  0.27%    10.66ms 0.022%  internal/sync.(*Mutex).unlockSlow
         0     0%  0.27%    10.66ms 0.022%  internal/sync.runtime_Semrelease
         0     0%  0.27%    10.66ms 0.022%  net.(*conn).SetReadDeadline
         0     0%  0.27%    10.66ms 0.022%  net.(*netFD).SetReadDeadline (inline)
         0     0%  0.27%    10.66ms 0.022%  net/http.(*connReader).abortPendingRead
    3.55ms 0.0073%  0.28%    10.66ms 0.022%  net/http.(*response).WriteHeader
    3.55ms 0.0073%  0.29%    10.66ms 0.022%  net/http.Header.Get
    7.11ms 0.015%   0.3%    10.66ms 0.022%  runtime.(*gcWork).tryStealSpan
         0     0%   0.3%    10.66ms 0.022%  runtime.gcMarkDone
         0     0%   0.3%    10.66ms 0.022%  runtime.goexit0
         0     0%   0.3%    10.66ms 0.022%  runtime.gopark
         0     0%   0.3%    10.66ms 0.022%  runtime.goparkunlock (inline)
         0     0%   0.3%    10.66ms 0.022%  runtime.makemap_small
         0     0%   0.3%    10.66ms 0.022%  runtime.memclrNoHeapPointersChunked
   10.66ms 0.022%  0.32%    10.66ms 0.022%  runtime.objptr.objIndex (inline)
         0     0%  0.32%    10.66ms 0.022%  runtime.reentersyscall
         0     0%  0.32%    10.66ms 0.022%  runtime.scanObjectSmall
    3.55ms 0.0073%  0.33%    10.66ms 0.022%  strings.Trim
         0     0%  0.33%    10.66ms 0.022%  sync.(*RWMutex).Unlock
         0     0%  0.33%    10.42ms 0.021%  runtime.callers
         0     0%  0.33%    10.42ms 0.021%  runtime.callers.func1
         0     0%  0.33%      -10ms 0.021%  context.removeChild
     -10ms 0.021%  0.31%      -10ms 0.021%  fmt.(*fmt).fmtInteger
     -10ms 0.021%  0.29%      -10ms 0.021%  fmt.(*pp).argNumber
     -10ms 0.021%  0.27%      -10ms 0.021%  fmt.parsenum (inline)
     -10ms 0.021%  0.25%      -10ms 0.021%  go.uber.org/zap.anyFieldC[go.shape.string].Any
         0     0%  0.25%      -10ms 0.021%  internal/abi.(*FuncType).In (inline)
     -10ms 0.021%  0.23%      -10ms 0.021%  internal/abi.(*FuncType).NumIn (inline)
     -10ms 0.021%  0.21%      -10ms 0.021%  internal/abi.(*FuncType).NumOut (inline)
         0     0%  0.21%      -10ms 0.021%  internal/abi.(*FuncType).Out (inline)
     -10ms 0.021%  0.19%      -10ms 0.021%  internal/abi.(*FuncType).OutSlice (inline)
     -10ms 0.021%  0.17%      -10ms 0.021%  internal/abi.(*Type).Uncommon (inline)
         0     0%  0.17%      -10ms 0.021%  internal/abi.Name.DataChecked (inline)
         0     0%  0.17%      -10ms 0.021%  internal/abi.Name.Tag
     -10ms 0.021%  0.15%      -10ms 0.021%  internal/abi.addChecked (inline)
         0     0%  0.15%      -10ms 0.021%  internal/poll.(*pollDesc).prepare
         0     0%  0.15%      -10ms 0.021%  internal/poll.(*pollDesc).prepareWrite (inline)
     -10ms 0.021%  0.12%      -10ms 0.021%  internal/poll.convertErr (inline)
         0     0%  0.12%      -10ms 0.021%  internal/runtime/maps.(*Map).Delete
         0     0%  0.12%      -10ms 0.021%  internal/runtime/maps.(*Map).deleteSmall
     -10ms 0.021%   0.1%      -10ms 0.021%  internal/runtime/maps.(*Map).directoryAt (inline)
         0     0%   0.1%      -10ms 0.021%  internal/runtime/maps.typedmemclr
     -10ms 0.021% 0.084%      -10ms 0.021%  net.(*TCPConn).Write
     -10ms 0.021% 0.063%      -10ms 0.021%  net/http.writeStatusLine
     -10ms 0.021% 0.042%      -10ms 0.021%  reflect.(*rtype).MethodByName
     -10ms 0.021% 0.022%      -10ms 0.021%  reflect.Value.String (inline)
     -10ms 0.021% 0.0013%      -10ms 0.021%  runtime.(*_panic).start
     -10ms 0.021% 0.019%      -10ms 0.021%  runtime.(*gcControllerState).releaseNextGCMarkWorker
     -10ms 0.021%  0.04%      -10ms 0.021%  runtime.(*gcWork).putObjFast (inline)
         0     0%  0.04%      -10ms 0.021%  runtime.(*mspan).heapBits
         0     0%  0.04%      -10ms 0.021%  runtime.(*mspan).initInlineMarkBits (inline)
     -10ms 0.021%  0.06%      -10ms 0.021%  runtime.(*randomEnum).position (inline)
     -10ms 0.021% 0.081%      -10ms 0.021%  runtime.(*semaRoot).queue
     -10ms 0.021%   0.1%      -10ms 0.021%  runtime.(*spanInlineMarkBits).init
     -10ms 0.021%  0.12%      -10ms 0.021%  runtime.(*stackScanState).getPtr
     -10ms 0.021%  0.14%      -10ms 0.021%  runtime.(*stkframe).argMapInternal
     -10ms 0.021%  0.16%      -10ms 0.021%  runtime.adjustctxt
     -10ms 0.021%  0.18%      -10ms 0.021%  runtime.bool2int (inline)
         0     0%  0.18%      -10ms 0.021%  runtime.casGToWaiting (inline)
         0     0%  0.18%      -10ms 0.021%  runtime.casGToWaitingForSuspendG
     -10ms 0.021%   0.2%      -10ms 0.021%  runtime.eqslice (inline)
     -10ms 0.021%  0.22%      -10ms 0.021%  runtime.forEachG
     -10ms 0.021%  0.25%      -10ms 0.021%  runtime.freeDeadSpanSPMCs
         0     0%  0.25%      -10ms 0.021%  runtime.gcResetMarkState
     -10ms 0.021%  0.27%      -10ms 0.021%  runtime.heapBitsInSpan (inline)
     -10ms 0.021%  0.29%      -10ms 0.021%  runtime.heapBitsSlice (inline)
         0     0%  0.29%      -10ms 0.021%  runtime.makeSpanClass (inline)
         0     0%  0.29%      -10ms 0.021%  runtime.mapdelete
     -10ms 0.021%  0.31%      -10ms 0.021%  runtime.mergeSummaries
         0     0%  0.31%      -10ms 0.021%  runtime.newAllocBits (inline)
     -10ms 0.021%  0.33%      -10ms 0.021%  runtime.pcdatastart (inline)
     -10ms 0.021%  0.35%      -10ms 0.021%  runtime.procUnpin (inline)
     -10ms 0.021%  0.37%      -10ms 0.021%  runtime.readgstatus (inline)
         0     0%  0.37%      -10ms 0.021%  runtime.releasep
         0     0%  0.37%      -10ms 0.021%  runtime.releasepNoTrace
         0     0%  0.37%      -10ms 0.021%  runtime.stkbucket
     -10ms 0.021%  0.39%      -10ms 0.021%  runtime.strhash
     -10ms 0.021%  0.41%      -10ms 0.021%  runtime.traceLocker.ok (inline)
         0     0%  0.41%      -10ms 0.021%  runtime.trygetfull
         0     0%  0.41%      -10ms 0.021%  runtime.typedmemclr
     -10ms 0.021%  0.43%      -10ms 0.021%  sync.(*RWMutex).RUnlock
         0     0%  0.43%      -10ms 0.021%  sync.runtime_procUnpin
     -10ms 0.021%  0.45%      -10ms 0.021%  sync/atomic.(*Pointer[go.shape.struct { sync.poolDequeue; sync.next sync/atomic.Pointer[sync.poolChainElt]; sync.prev sync/atomic.Pointer[sync.poolChainElt] }]).Load (inline)
     -10ms 0.021%  0.47%      -10ms 0.021%  sync/atomic.(*Uint32).Load (inline)
     -10ms 0.021%  0.49%      -10ms 0.021%  text/template.(*Template).AddParseTree
         0     0%  0.49%      -10ms 0.021%  text/template.(*state).evalCall.func1 (inline)
         0     0%  0.49%      -10ms 0.021%  text/template.(*state).evalString
     -10ms 0.021%  0.51%      -10ms 0.021%  text/template.goodFunc
     -10ms 0.021%  0.53%      -10ms 0.021%  text/template.isTrue
         0     0%  0.53%      -10ms 0.021%  text/template/parse.(*CommandNode).append (inline)
         0     0%  0.53%      -10ms 0.021%  text/template/parse.(*Tree).add
    9.76ms  0.02%  0.51%     9.76ms  0.02%  runtime.nextFreeFast (inline)
     -10ms 0.021%  0.53%    -9.58ms  0.02%  sync.(*Pool).Get
     -20ms 0.041%  0.57%    -9.34ms 0.019%  internal/sync.(*Mutex).Unlock (inline)
         0     0%  0.57%    -9.34ms 0.019%  sync.(*Mutex).Unlock (partial-inline)
         0     0%  0.57%     8.98ms 0.018%  runtime.startm
    3.55ms 0.0073%  0.57%    -8.67ms 0.018%  net/http.(*chunkWriter).close
         0     0%  0.57%    -8.67ms 0.018%  runtime.(*inlineUnwinder).resolveInternal (inline)
         0     0%  0.57%    -8.67ms 0.018%  runtime.newInlineUnwinder
         0     0%  0.57%    -8.67ms 0.018%  runtime.pcdatavalue1
    8.43ms 0.017%  0.55%     8.43ms 0.017%  reflect.Value.Field
         0     0%  0.55%     8.43ms 0.017%  runtime.gcMarkTermination.func4
   -2.89ms 0.0059%  0.56%     8.43ms 0.017%  runtime.nilinterhash
         0     0%  0.56%     8.43ms 0.017%  time.Time.Format
    4.22ms 0.0087%  0.55%     7.77ms 0.016%  encoding/json.(*scanner).pushParseState
    7.77ms 0.016%  0.53%     7.77ms 0.016%  internal/strconv.atof64exact
         0     0%  0.53%     7.77ms 0.016%  io.CopyN
         0     0%  0.53%     7.77ms 0.016%  io.LimitReader (inline)
    7.77ms 0.016%  0.52%     7.77ms 0.016%  runtime.alignDown (inline)
    7.77ms 0.016%   0.5%     7.77ms 0.016%  runtime.getGCMask (inline)
   42.60ms 0.088%  0.41%    -7.76ms 0.016%  runtime.scanObject
         0     0%  0.41%     7.11ms 0.015%  encoding/json.newEncodeState
    3.55ms 0.0073%   0.4%     7.11ms 0.015%  encoding/json.state0
         0     0%   0.4%     7.11ms 0.015%  github.com/paveltovchigrechko/metrics-service/internal/handler.(*AppHandler).UpdateEndpoint
         0     0%   0.4%     7.11ms 0.015%  github.com/paveltovchigrechko/metrics-service/internal/repository.(*MemStorage).GetMetrics
         0     0%   0.4%     7.11ms 0.015%  go.uber.org/zap/buffer.(*Buffer).AppendInt (inline)
         0     0%   0.4%     7.11ms 0.015%  go.uber.org/zap/internal/stacktrace.(*Stack).Next (inline)
         0     0%   0.4%     7.11ms 0.015%  go.uber.org/zap/zapcore.(*jsonEncoder).AddInt64
         0     0%   0.4%     7.11ms 0.015%  go.uber.org/zap/zapcore.(*jsonEncoder).addKey
    7.11ms 0.015%  0.39%     7.11ms 0.015%  internal/abi.NoEscape
         0     0%  0.39%     7.11ms 0.015%  internal/runtime/maps.NewMap
    7.11ms 0.015%  0.38%     7.11ms 0.015%  internal/strconv.special
         0     0%  0.38%     7.11ms 0.015%  net/http.readTransfer
         0     0%  0.38%     7.11ms 0.015%  net/textproto.CanonicalMIMEHeaderKey
         0     0%  0.38%     7.11ms 0.015%  net/textproto.MIMEHeader.Get
    3.55ms 0.0073%  0.37%     7.11ms 0.015%  net/textproto.validHeaderFieldByte (inline)
         0     0%  0.37%     7.11ms 0.015%  net/url.ParseRequestURI
         0     0%  0.37%     7.11ms 0.015%  net/url.parse
    7.11ms 0.015%  0.35%     7.11ms 0.015%  reflect.Value.SetFloat
         0     0%  0.35%     7.11ms 0.015%  runtime.(*Frames).Next
    7.11ms 0.015%  0.34%     7.11ms 0.015%  runtime.(*activeSweep).end
    7.11ms 0.015%  0.32%     7.11ms 0.015%  runtime.(*consistentHeapStats).release
         0     0%  0.32%     7.11ms 0.015%  runtime.(*gcControllerState).update
    3.55ms 0.0073%  0.32%     7.11ms 0.015%  runtime.(*gcWork).tryGetSpan
    7.11ms 0.015%   0.3%     7.11ms 0.015%  runtime.(*mspan).moveInlineMarks
    7.11ms 0.015%  0.29%     7.11ms 0.015%  runtime.(*mspan).typePointersOfType
         0     0%  0.29%     7.11ms 0.015%  runtime.(*pageAlloc).allocToCache
    7.11ms 0.015%  0.27%     7.11ms 0.015%  runtime.(*scavChunkData).free
         0     0%  0.27%     7.11ms 0.015%  runtime.(*scavengeIndex).free
    7.11ms 0.015%  0.26%     7.11ms 0.015%  runtime.(*spanQueue).refill
         0     0%  0.26%     7.11ms 0.015%  runtime.(*spanQueue).steal
         0     0%  0.26%     7.11ms 0.015%  runtime.(*stackScanState).buildIndex (inline)
         0     0%  0.26%     7.11ms 0.015%  runtime.asyncPreempt2.func1
    7.11ms 0.015%  0.24%     7.11ms 0.015%  runtime.binarySearchTree
         0     0%  0.24%     7.11ms 0.015%  runtime.entersyscallWakeSysmon
         0     0%  0.24%     7.11ms 0.015%  runtime.forEachP (inline)
    3.55ms 0.0073%  0.24%     7.11ms 0.015%  runtime.funcline1
    7.11ms 0.015%  0.22%     7.11ms 0.015%  runtime.gcWriteBarrier2
         0     0%  0.22%     7.11ms 0.015%  runtime.goready (inline)
         0     0%  0.22%     7.11ms 0.015%  runtime.makemap
         0     0%  0.22%     7.11ms 0.015%  runtime.mapIterStart
    7.11ms 0.015%  0.21%     7.11ms 0.015%  runtime.markBits.isMarked (inline)
    7.11ms 0.015%  0.19%     7.11ms 0.015%  runtime.memhash64
    7.11ms 0.015%  0.18%     7.11ms 0.015%  runtime.newArenaMayUnlock
    7.11ms 0.015%  0.16%     7.11ms 0.015%  strings.trimLeftByte (inline)
         0     0%  0.16%     7.11ms 0.015%  sync.runtime_SemacquireRWMutexR
         0     0%  0.16%     7.11ms 0.015%  time.Until
    7.05ms 0.015%  0.15%     7.05ms 0.015%  runtime.memmove
    6.87ms 0.014%  0.13%     6.87ms 0.014%  runtime.memequal
         0     0%  0.13%    -6.45ms 0.013%  context.(*cancelCtx).cancel
         0     0%  0.13%    -6.45ms 0.013%  context.WithCancel.func1
    3.55ms 0.0073%  0.13%    -6.45ms 0.013%  go.uber.org/zap.Any
         0     0%  0.13%    -6.45ms 0.013%  go.uber.org/zap/zapcore.EntryCaller.TrimmedPath
   -6.45ms 0.013%  0.14%    -6.45ms 0.013%  internal/bytealg.IndexByteString
     -10ms 0.021%  0.16%    -6.45ms 0.013%  internal/runtime/atomic.(*Bool).Load (inline)
   -6.45ms 0.013%  0.17%    -6.45ms 0.013%  reflect.(*rtype).Kind (inline)
   -6.45ms 0.013%  0.19%    -6.45ms 0.013%  reflect.Value.Index
   -6.45ms 0.013%   0.2%    -6.45ms 0.013%  reflect.Value.SetLen
   -6.45ms 0.013%  0.21%    -6.45ms 0.013%  runtime.(*atomicHeadTailIndex).incTail
         0     0%  0.21%    -6.45ms 0.013%  runtime.(*gcControllerState).heapGoalInternal
     -10ms 0.021%  0.24%    -6.45ms 0.013%  runtime.(*gcControllerState).memoryLimitHeapGoal
   -6.45ms 0.013%  0.25%    -6.45ms 0.013%  runtime.(*lfstack).pop
         0     0%  0.25%    -6.45ms 0.013%  runtime.(*m).clearAllpSnapshot
   -6.45ms 0.013%  0.26%    -6.45ms 0.013%  runtime.(*mspan).countAlloc (inline)
         0     0%  0.26%    -6.45ms 0.013%  runtime.(*pageAlloc).allocRange
     -10ms 0.021%  0.28%    -6.45ms 0.013%  runtime.(*pageAlloc).scavenge
         0     0%  0.28%    -6.45ms 0.013%  runtime.gcMarkTermination.func4.1
    3.55ms 0.0073%  0.27%    -6.45ms 0.013%  runtime.isShrinkStackSafe (inline)
   -6.45ms 0.013%  0.29%    -6.45ms 0.013%  runtime.libcCall
   -6.45ms 0.013%   0.3%    -6.45ms 0.013%  runtime.pidleget
         0     0%   0.3%    -6.45ms 0.013%  runtime.pidlegetSpinning
   -6.45ms 0.013%  0.31%    -6.45ms 0.013%  runtime.readUintptr (inline)
   -6.45ms 0.013%  0.33%    -6.45ms 0.013%  runtime.save_g
   -6.45ms 0.013%  0.34%    -6.45ms 0.013%  sync/atomic.(*Int32).Add (inline)
   -6.45ms 0.013%  0.35%    -6.45ms 0.013%  time.nextStdChunk
         0     0%  0.35%    -5.78ms 0.012%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
         0     0%  0.35%    -5.78ms 0.012%  runtime.(*gcControllerState).findRunnableGCWorker
   21.99ms 0.045%  0.31%     5.30ms 0.011%  reflect.Value.Set
   -6.45ms 0.013%  0.32%    -4.70ms 0.0097%  runtime.(*unwinder).resolveInternal
     -10ms 0.021%  0.34%     4.40ms 0.009%  runtime.bgsweep
    4.22ms 0.0087%  0.33%     4.22ms 0.0087%  reflect.flag.kind (inline)
    4.22ms 0.0087%  0.33%     4.22ms 0.0087%  runtime.(*consistentHeapStats).acquire
    0.66ms 0.0014%  0.32%     4.22ms 0.0087%  runtime.(*mspan).markBitsForIndex (inline)
         0     0%  0.32%     4.22ms 0.0087%  runtime.(*pageAlloc).free
    4.22ms 0.0087%  0.32%     4.22ms 0.0087%  runtime.(*pallocBits).summarize
    3.55ms 0.0073%  0.31%     4.22ms 0.0087%  runtime.markrootSpans
         0     0%  0.31%     3.98ms 0.0082%  runtime.startTheWorldWithSema
    3.55ms 0.0073%   0.3%     3.55ms 0.0073%  bytes.TrimLeft
         0     0%   0.3%     3.55ms 0.0073%  context.(*cancelCtx).propagateCancel
         0     0%   0.3%     3.55ms 0.0073%  context.WithCancel
         0     0%   0.3%     3.55ms 0.0073%  context.withCancel (inline)
         0     0%   0.3%     3.55ms 0.0073%  encoding/json.(*decodeState).init (inline)
    3.55ms 0.0073%  0.29%     3.55ms 0.0073%  encoding/json.stateBeginValueOrEmpty
    3.55ms 0.0073%  0.29%     3.55ms 0.0073%  fmt.(*buffer).writeString (inline)
         0     0%  0.29%     3.55ms 0.0073%  github.com/go-chi/chi/v5.(*Context).RoutePattern
         0     0%  0.29%     3.55ms 0.0073%  github.com/paveltovchigrechko/metrics-service/internal/repository.(*MemStorage).SaveMetrics
         0     0%  0.29%     3.55ms 0.0073%  go.uber.org/zap/zapcore.(*jsonEncoder).AddBool
         0     0%  0.29%     3.55ms 0.0073%  go.uber.org/zap/zapcore.(*jsonEncoder).AddDuration
         0     0%  0.29%     3.55ms 0.0073%  go.uber.org/zap/zapcore.(*jsonEncoder).AppendDuration
         0     0%  0.29%     3.55ms 0.0073%  go.uber.org/zap/zapcore.(*jsonEncoder).AppendInt64
         0     0%  0.29%     3.55ms 0.0073%  go.uber.org/zap/zapcore.(*jsonEncoder).Clone
         0     0%  0.29%     3.55ms 0.0073%  go.uber.org/zap/zapcore.(*jsonEncoder).clone
         0     0%  0.29%     3.55ms 0.0073%  go.uber.org/zap/zapcore.CapitalLevelEncoder
         0     0%  0.29%     3.55ms 0.0073%  go.uber.org/zap/zapcore.StringDurationEncoder
    3.55ms 0.0073%  0.28%     3.55ms 0.0073%  go.uber.org/zap/zapcore.init.New[go.shape.*uint8].func8
         0     0%  0.28%     3.55ms 0.0073%  go.uber.org/zap/zapcore.systemClock.Now
    3.55ms 0.0073%  0.27%     3.55ms 0.0073%  internal/abi.(*MapType).IndirectKey (inline)
    3.55ms 0.0073%  0.26%     3.55ms 0.0073%  internal/chacha8rand.block
    3.55ms 0.0073%  0.26%     3.55ms 0.0073%  internal/poll.runtime_pollSetDeadline
    3.55ms 0.0073%  0.25%     3.55ms 0.0073%  internal/runtime/atomic.(*Uint32).Load (inline)
    3.55ms 0.0073%  0.24%     3.55ms 0.0073%  internal/runtime/atomic.(*Uint8).Load (inline)
    3.55ms 0.0073%  0.24%     3.55ms 0.0073%  internal/runtime/maps.(*Map).Used (inline)
    3.55ms 0.0073%  0.23%     3.55ms 0.0073%  internal/runtime/maps.(*Map).putSlotSmall
    3.55ms 0.0073%  0.22%     3.55ms 0.0073%  internal/runtime/maps.(*groupReference).elem (inline)
         0     0%  0.22%     3.55ms 0.0073%  internal/runtime/maps.ctrlGroup.matchEmpty (inline)
    3.55ms 0.0073%  0.21%     3.55ms 0.0073%  internal/runtime/maps.ctrlGroupMatchEmpty (inline)
    3.55ms 0.0073%  0.21%     3.55ms 0.0073%  internal/strconv.small (inline)
         0     0%  0.21%     3.55ms 0.0073%  internal/stringslite.Cut
         0     0%  0.21%     3.55ms 0.0073%  internal/stringslite.Index
         0     0%  0.21%     3.55ms 0.0073%  internal/stringslite.IndexByte (inline)
         0     0%  0.21%     3.55ms 0.0073%  internal/sync.runtime_nanotime
         0     0%  0.21%     3.55ms 0.0073%  io.Copy (inline)
         0     0%  0.21%     3.55ms 0.0073%  io.MultiReader (inline)
         0     0%  0.21%     3.55ms 0.0073%  io.ReadAtLeast
         0     0%  0.21%     3.55ms 0.0073%  io.ReadFull (inline)
         0     0%  0.21%     3.55ms 0.0073%  io.copyBuffer
         0     0%  0.21%     3.55ms 0.0073%  io.discard.ReadFrom
         0     0%  0.21%     3.55ms 0.0073%  io.init.func1
         0     0%  0.21%     3.55ms 0.0073%  net/http.Header.Clone (inline)
         0     0%  0.21%     3.55ms 0.0073%  net/http.Header.get (inline)
         0     0%  0.21%     3.55ms 0.0073%  net/http.extraHeader.Write
         0     0%  0.21%     3.55ms 0.0073%  net/http.init.func15
         0     0%  0.21%     3.55ms 0.0073%  net/http.newBufioWriterSize
         0     0%  0.21%     3.55ms 0.0073%  net/http.parseRequestLine
         0     0%  0.21%     3.55ms 0.0073%  net/http.putBufioWriter
    3.55ms 0.0073%   0.2%     3.55ms 0.0073%  net/url.stringContainsCTLByte (inline)
    3.55ms 0.0073%  0.19%     3.55ms 0.0073%  reflect.(*rtype).uncommon (inline)
    3.55ms 0.0073%  0.18%     3.55ms 0.0073%  reflect.Value.Cap (inline)
    3.55ms 0.0073%  0.18%     3.55ms 0.0073%  reflect.Value.OverflowInt (inline)
    3.55ms 0.0073%  0.17%     3.55ms 0.0073%  reflect.Value.SetInt
    3.55ms 0.0073%  0.16%     3.55ms 0.0073%  reflect.flag.mustBeAssignable (inline)
    3.55ms 0.0073%  0.15%     3.55ms 0.0073%  reflect.flag.mustBeExported (inline)
         0     0%  0.15%     3.55ms 0.0073%  runtime.(*_panic).nextDefer
    3.55ms 0.0073%  0.15%     3.55ms 0.0073%  runtime.(*gcBits).bytep (inline)
         0     0%  0.15%     3.55ms 0.0073%  runtime.(*gcCPULimiterState).update
         0     0%  0.15%     3.55ms 0.0073%  runtime.(*gcCPULimiterState).updateLocked
    3.55ms 0.0073%  0.14%     3.55ms 0.0073%  runtime.(*gcControllerState).addScannableStack (inline)
         0     0%  0.14%     3.55ms 0.0073%  runtime.(*gcControllerState).assignWaitingGCWorker
         0     0%  0.14%     3.55ms 0.0073%  runtime.(*gcWork).init
         0     0%  0.14%     3.55ms 0.0073%  runtime.(*inlineUnwinder).next
    3.55ms 0.0073%  0.13%     3.55ms 0.0073%  runtime.(*listHeadManual).pop
    3.55ms 0.0073%  0.13%     3.55ms 0.0073%  runtime.(*mLockProfile).store (inline)
         0     0%  0.13%     3.55ms 0.0073%  runtime.(*mheap).freeManual
    3.55ms 0.0073%  0.12%     3.55ms 0.0073%  runtime.(*mspan).scannedBitsForIndex (inline)
         0     0%  0.12%     3.55ms 0.0073%  runtime.(*pageAlloc).find
    3.55ms 0.0073%  0.11%     3.55ms 0.0073%  runtime.(*pageAlloc).find.func1
         0     0%  0.11%     3.55ms 0.0073%  runtime.(*pallocData).findScavengeCandidate
    3.55ms 0.0073%   0.1%     3.55ms 0.0073%  runtime.(*spanQueue).put
    3.55ms 0.0073% 0.096%     3.55ms 0.0073%  runtime.(*stackScanState).findObject (inline)
    3.55ms 0.0073% 0.089%     3.55ms 0.0073%  runtime.(*sysMemStat).load (inline)
         0     0% 0.089%     3.55ms 0.0073%  runtime.(*timer).unlockAndRun
         0     0% 0.089%     3.55ms 0.0073%  runtime.(*timers).run
         0     0% 0.089%     3.55ms 0.0073%  runtime.acquireSudog
    3.55ms 0.0073% 0.082%     3.55ms 0.0073%  runtime.atomicwb
    3.55ms 0.0073% 0.074%     3.55ms 0.0073%  runtime.canPreemptM (inline)
         0     0% 0.074%     3.55ms 0.0073%  runtime.deductSweepCredit
    3.55ms 0.0073% 0.067%     3.55ms 0.0073%  runtime.fillAligned
         0     0% 0.067%     3.55ms 0.0073%  runtime.freeSomeWbufs.func1
    3.55ms 0.0073%  0.06%     3.55ms 0.0073%  runtime.funcdata (inline)
         0     0%  0.06%     3.55ms 0.0073%  runtime.gcMarkTermination
    3.55ms 0.0073% 0.052%     3.55ms 0.0073%  runtime.gcScanCleanup (inline)
    3.55ms 0.0073% 0.045%     3.55ms 0.0073%  runtime.gcWriteBarrier1
         0     0% 0.045%     3.55ms 0.0073%  runtime.gdestroy
    3.55ms 0.0073% 0.038%     3.55ms 0.0073%  runtime.getcallerfp
         0     0% 0.038%     3.55ms 0.0073%  runtime.gfget
         0     0% 0.038%     3.55ms 0.0073%  runtime.gfget.func2
         0     0% 0.038%     3.55ms 0.0073%  runtime.goroutineReady
         0     0% 0.038%     3.55ms 0.0073%  runtime.goroutineReady.goready.func1
         0     0% 0.038%     3.55ms 0.0073%  runtime.goschedIfBusy
         0     0% 0.038%     3.55ms 0.0073%  runtime.makeslicecopy
         0     0% 0.038%     3.55ms 0.0073%  runtime.mapassign
         0     0% 0.038%     3.55ms 0.0073%  runtime.mget (inline)
         0     0% 0.038%     3.55ms 0.0073%  runtime.netpollBreak (inline)
         0     0% 0.038%     3.55ms 0.0073%  runtime.parkunlock_c
    3.55ms 0.0073%  0.03%     3.55ms 0.0073%  runtime.popDefer
    3.55ms 0.0073% 0.023%     3.55ms 0.0073%  runtime.procPin (inline)
    3.55ms 0.0073% 0.016%     3.55ms 0.0073%  runtime.pthread_mutex_unlock
    3.55ms 0.0073% 0.0086%     3.55ms 0.0073%  runtime.putempty
         0     0% 0.0086%     3.55ms 0.0073%  runtime.readyWithTime
         0     0% 0.0086%     3.55ms 0.0073%  runtime.semrelease (inline)
         0     0% 0.0086%     3.55ms 0.0073%  runtime.stackalloc
         0     0% 0.0086%     3.55ms 0.0073%  runtime.stackcacherefill
         0     0% 0.0086%     3.55ms 0.0073%  runtime.stackpoolalloc
         0     0% 0.0086%     3.55ms 0.0073%  runtime.traceAcquire (inline)
    3.55ms 0.0073% 0.0012%     3.55ms 0.0073%  runtime.traceEnabled (inline)
    3.55ms 0.0073% 0.0061%     3.55ms 0.0073%  runtime.traceShuttingDown (inline)
         0     0% 0.0061%     3.55ms 0.0073%  runtime.wakeNetpoll
         0     0% 0.0061%     3.55ms 0.0073%  strings.Cut (inline)
    3.55ms 0.0073% 0.013%     3.55ms 0.0073%  strings.Join
         0     0% 0.013%     3.55ms 0.0073%  sync.(*Cond).Wait
    3.55ms 0.0073% 0.021%     3.55ms 0.0073%  sync.(*poolDequeue).popTail
         0     0% 0.021%     3.55ms 0.0073%  sync.runtime_notifyListWait
         0     0% 0.021%     3.55ms 0.0073%  sync/atomic.(*Value).Store
         0     0% 0.021%     3.55ms 0.0073%  sync/atomic.StorePointer
         0     0% 0.021%     3.55ms 0.0073%  sync/atomic.runtime_procPin
    3.55ms 0.0073% 0.028%     3.55ms 0.0073%  time.(*Time).sec (inline)
         0     0% 0.028%     3.55ms 0.0073%  time.(*Time).unixSec (inline)
         0     0% 0.028%     3.55ms 0.0073%  time.Duration.String (inline)
         0     0% 0.028%     3.55ms 0.0073%  time.Since
    3.55ms 0.0073% 0.035%     3.55ms 0.0073%  time.Time.Add
         0     0% 0.035%     3.55ms 0.0073%  time.Time.Sub
         0     0% 0.035%     3.55ms 0.0073%  time.Time.locabs
    3.55ms 0.0073% 0.043%     3.55ms 0.0073%  time.absCentury.year (inline)
         0     0% 0.043%     3.55ms 0.0073%  time.absDays.date
         0     0% 0.043%     3.55ms 0.0073%  time.runtimeNano
    3.55ms 0.0073%  0.05%     3.55ms 0.0073%  vendor/golang.org/x/net/http/httpguts.ValidHeaderFieldName (inline)
         0     0%  0.05%     3.31ms 0.0068%  runtime.tracebackPCs
    3.55ms 0.0073% 0.057%    -3.13ms 0.0064%  runtime.(*mheap).alloc
         0     0% 0.057%    -2.89ms 0.0059%  go.uber.org/zap/zapcore.getCheckedEntry
   -2.89ms 0.0059% 0.051%    -2.89ms 0.0059%  internal/abi.(*Type).Elem (inline)
   -2.89ms 0.0059% 0.045%    -2.89ms 0.0059%  internal/abi.(*Type).IsDirectIface (inline)
   -2.89ms 0.0059% 0.039%    -2.89ms 0.0059%  internal/abi.(*Type).Pointers (inline)
         0     0% 0.039%    -2.89ms 0.0059%  internal/runtime/maps.(*Map).growToSmall
   -2.89ms 0.0059% 0.033%    -2.89ms 0.0059%  internal/runtime/maps.(*ctrlGroup).get (inline)
         0     0% 0.033%    -2.89ms 0.0059%  internal/runtime/maps.newGroups (inline)
     -10ms 0.021% 0.013%    -2.89ms 0.0059%  internal/runtime/maps.newarray
   -2.89ms 0.0059% 0.0069%    -2.89ms 0.0059%  internal/strconv.pow10 (inline)
    7.11ms 0.015% 0.022%    -2.89ms 0.0059%  internal/sync.(*HashTrieMap[go.shape.interface {},go.shape.interface {}]).init (inline)
         0     0% 0.022%    -2.89ms 0.0059%  net/textproto.NewReader (inline)
   -2.89ms 0.0059% 0.016%    -2.89ms 0.0059%  reflect.flag.ro (inline)
    3.55ms 0.0073% 0.023%    -2.89ms 0.0059%  runtime.(*gcControllerState).trigger
         0     0% 0.023%    -2.89ms 0.0059%  runtime.(*gcWork).tryGetObj
   -2.89ms 0.0059% 0.017%    -2.89ms 0.0059%  runtime.(*mspan).nextFreeIndex
         0     0% 0.017%    -2.89ms 0.0059%  runtime.(*pageAlloc).alloc
         0     0% 0.017%    -2.89ms 0.0059%  runtime.(*scavengerState).init.func2
         0     0% 0.017%    -2.89ms 0.0059%  runtime.(*scavengerState).run
         0     0% 0.017%    -2.89ms 0.0059%  runtime.bgscavenge
         0     0% 0.017%    -2.89ms 0.0059%  runtime.gcNextMarkRoot
         0     0% 0.017%    -2.89ms 0.0059%  runtime.gcScanFinalizer
         0     0% 0.017%    -2.89ms 0.0059%  runtime.gcTrigger.test
         0     0% 0.017%    -2.89ms 0.0059%  runtime.handoffp
         0     0% 0.017%    -2.89ms 0.0059%  runtime.mapassign_faststr
   -2.89ms 0.0059% 0.011%    -2.89ms 0.0059%  runtime.pageIndexOf (inline)
         0     0% 0.011%    -2.23ms 0.0046%  runtime.(*mheap).freeSpanLocked
   -2.23ms 0.0046% 0.0064%    -2.23ms 0.0046%  runtime.(*moduledata).textAddr
    3.55ms 0.0073% 0.014%    -2.23ms 0.0046%  runtime.(*pageAlloc).update
         0     0% 0.014%    -2.23ms 0.0046%  runtime.funcInfo.entry (inline)
  -12.89ms 0.027% 0.013%     1.99ms 0.0041%  go.uber.org/zap.(*SugaredLogger).sweetenFields
         0     0% 0.013%     1.75ms 0.0036%  runtime.funcspdelta (inline)
    1.33ms 0.0027%  0.01%     1.33ms 0.0027%  runtime.findfunc
         0     0%  0.01%     1.33ms 0.0027%  runtime.netpollgoready.goready.func1
         0     0%  0.01%     0.66ms 0.0014%  go.uber.org/zap/zapcore.(*CheckedEntry).AddCore (inline)
         0     0%  0.01%     0.66ms 0.0014%  go.uber.org/zap/zapcore.(*ioCore).Check
    0.66ms 0.0014% 0.0087%     0.66ms 0.0014%  internal/runtime/atomic.(*Uint32).Add (inline)
         0     0% 0.0087%     0.66ms 0.0014%  net/http.newTextprotoReader
    0.66ms 0.0014% 0.0073%     0.66ms 0.0014%  runtime.(*mspan).base (inline)
    3.55ms 0.0073% 3.4e-07%     0.66ms 0.0014%  runtime.(*spanSet).push
