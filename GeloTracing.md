Gelo has a simple mechanism for tracing the execution of Gelo programs. It is global to all VMs. The only granularity is in what traces to display.

There are four kinds of traces, each have an associated constant in the `gelo` package and a mnemonic displayed in the TraceFormat:
|**Kind**|**Constant**|**Mnemonic code**|**Description**|
|:-------|:-----------|:----------------|:--------------|
|Parser  |`Parser_trace`|`P`              |Traces how the parser parses the lines in a quote|
|System  |`System_trace`|`S`              |Traces system level events, such as a new VM being created|
|Runtime |`Runtime_trace`|`R`              |Traces runtime events, such as the result of rewriting a line|
|Alien   |`Alien_trace`|`X`              |Traces emanating from [alien commands](GeloTypes#Alien.md)|

In addition to the above, there is a convienence constant `All_traces = Alien_trace | Runtime_trace | System_trace | Parser_trace`.

To turn traces on and off there are two functions:

  * `TraceOn(_trace) _trace`
  * `TraceOff(_trace) _trace`

The unexported `_trace` type is the type of the constants above. The functions only toggle the trace you give it, so if Alien trace is on and you call `TraceOn(Runtime_trace)` then both will be set and `Alien_trace | Runtime_trace` will be returned.

By default, there is no tracer set and all traces are lost. A tracer is a [Port](GeloTypes#Port.md) and can be set by `SetTracer(Port) Port`. It returns the previous tracer or `nil` if there wasn't one.