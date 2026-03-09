# notes for myself

The renderer (and `ReadSamples`) is supposed to work like this:

* drain the internal buffer
* if the buffer is empty:
  * while sample quota is not met:
    * check if the queue is empty and return EOF if it is
    * pop all notes to be rendered up to the requested time (i.e. # of samples) in seconds
    * for each note (concurrently):
      * check cache and invoke the resampler
      * invoke the concatenator (wavtool) with the internal buffer + resampled note
      * write the result to the internal buffer
