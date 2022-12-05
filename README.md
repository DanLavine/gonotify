# gonotify

Go Notify can be used to work along side other data types that want to trigger information is available.

## Feaure

Doesn't spin up a large number of Gorotines to trigger that data is available. Uses just 1 to manage
all the notifications. Also doesn't use a buffered channel since there isn't always a known size for
how large that should be.

## Use Cases

Perhaps you have a in memory buffer of a type:
```
genericBuffer := [][]byte{}
```

This could be some sort of generic queue where you want to gurantee message order. But there could be any number of
Producers and Consumers at once. We don't know if there will be more write making this grow very large. Or more
Reads, ensuring this stays small.

On a Write to the `genericBuffer` we simply want to add the entire body and trigger there is a read ready:
```
... // do some proper logic to lock things
genericBuffer = append(genericBuffer, []byte(...))

go func() {
  // trigger a client that there is data to read
  readReady <- chan struct{}
}()
```

Then on each Read, we want to consume the first index of the buffer:
```
... // Do some proper logic to lock things
dataToReturn := genericBuffer[0]
... // do some proper logic to remove the index from the slice and set genericBuffer to be smaller
return dataToReturn
```

This could have a lot of goroutines to know that there is data to be read. Which is kind of a pain
to debug if things crash and print all stack traces. Its also hard to ensure that all those routines
are properly drained on a shutdown which is annoying

Instead we could use the Notifier to know when there is data written to + read from our generic buffer
and have a few pieces to ForceShutdown if things are taking to long. I.E. there are no consumers to read
messages so our buffer will always have data. If thats the case, just force a shutdown!
