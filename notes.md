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

## Resampler

```text
Voice resampler tool Ver14.02r
Copyright (C) 2008-2012 Ameya/Ayame
Usage is ...
resampler.exe <input wavfile> <output file> <pitch_percent> <velocity> [<flags> [<offset> <length_require> [<fixed length> [<end_blank> [<volume> [<modulation> [<pich bend>...]]]]]]]
flags:
    N : No formant filter
    G : (Re)Generate frequency list
    T : Export text frequency list
    B : Bressiness parameter. B0..B100, default B50
ex :
resampler.exe infile.wav outfile.wav 120 100 GB60
```

## Wavtool

```text
wavtool2 <outfile> <infile> offset length p1 p2 p3 v1 v2 v3 v4 ovr p4 p5 v5
```
