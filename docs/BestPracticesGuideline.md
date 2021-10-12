# Azure Defender K8S In Cluster Defense Best Practices Guideline

## Table of Contents

TBD

## Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md#specifying-map-capacity-hints)
- [Don’t just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
- [CodeReviewComments - golang/go](https://github.com/golang/go/wiki/CodeReviewComments)

## Guideline - Best Practices

### Globals

- **Avoid mutable globals**
- **Enums**:
    - Start Enums at One -The standard way of introducing enumerations in Go is to declare a custom type and a const
      group with iota. Since variables have a 0 default value, you should usually start your enums on a non-zero value.
    - There are cases where using the zero value makes sense, for example when the zero value case is the desirable
      default behavior.

### Errors

- TODO - add best practices for how to work with errors (wrap error, fmt, structs , etc...)

- **Wrap error**:
    - it is recommended to add context where possible so that instead of a vague error such as "connection refused", you
      get more useful errors such as "call service foo: connection refused".
- **Reduce scope of variables** – e.g.:
  ```go
    if err := ioutil.WriteFile(name, data, 0644); err != nil { 
    ```


- **Don't panic**!
    - Code running in production must avoid panics - If an error occurs, the function must return an error and allow the
      caller to decide how to handle it.
    - Panic/recover is not an error handling strategy. A program must panic only when something irrecoverable happens
      such as a nil dereference.

- **Exit programs**:
    - Call one of os.Exit or log.Fatal* only in main(). All other functions should return errors to signal failure.

### Casting

- The single return value form of a type assertion will panic on an incorrect type. Therefore, always use the "comma ok"
  idiom:
  ```go
    numAsString, ok := num.(string)
    if !ok {
        // Handle this error gracefully
    }  
    ```

### Performance

- **Use strconv over fmt** - when converting primitives to/from strings, strconv is faster than fmt.
- **Specifying map capacity approximated size** e.g.:
  ```go
    make(map[T1]T2, approximatedSize)
    ```

### Channels

- **Always close channel when done**
- Before reading from channel check that the channel is still open.
    ```go
    j, channelOpen := <-channel
            if !channelOpen {
                ...
            } else {
                ...
            }
    ```
