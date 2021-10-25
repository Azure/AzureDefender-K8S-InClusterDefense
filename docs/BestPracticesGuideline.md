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

- When you create structure that implements interface, cast the strcut to the interface. write the casting above the
  interface:
  ```go
  package user
    
  type IPerson interface{
  GetId() string
    }
  
  type IUser interface{
  GetUserName() string
    }
    
  // User implements IUser interface
  var _ IUser = (*User)(nil)
  
  // User implements IPerson interface
  var _ IPerson = (*User)(nil)
  type User struct{
  Id string
  UserName string
  }
  
  func (user *User) GetId() string {return user.Id}
  func (user *User) GetUserName() string{return user.UserName}
    ```

### Performance

- **Use strconv over fmt** - when converting primitives to/from strings, strconv is faster than fmt.
- **Specifying map capacity approximated size** e.g.:
  ```go
    make(map[T1]T2, approximatedSize)
    ```

### Channels

- Before reading from a channel check that the channel is still open (reading from a closed channel return the zero value for the channel's type without blocking )
    ```go
    data, channelOpen := <-channel
            if !channelOpen {
                ...
            } else {
                ...
            }
    ```
  
- Sending data and error to channel: Because it is possible to send to a channel only one object, and it is very common to return from a function 2 values - data and an error, we created a struct for wrapping data and error into one struct.
  - Use utils\channel_data_wrapper - a struct that hold 2 members: data and error.
  - Enable sending and receiving data and error throw channels.
  - How to use:
    - Let f be a function that returns data of type *mystruct and error
      ```go
      func f() (*mystruct, error)
      ```
      - Create channel:
      ```go
        myChannel := make(chan *utils.ChannelDataWrapper)
        ```
      - Send to channel:
      ```go
        myChannel <- utils.NewChannelDataWrapper(f())
        ```
      - Read from channel:
      ```go
        channelDataWrapper := <- myChannel
        if channelDataWrapper.Err != nil {
            ...
          } else{
            data, ok = channelDataWrapper.DataWrapper.(*mystruct)
            if !ok{
                ...
            }
            ...
          }
        ```
      
### Deployment

- Containers:
  - TODO : Add securityContext section.
  - Never use "latest" tag:
    ```yaml
    containers:
          - name: {{.Values.AzDProxy.prefixResourceDeployment}}-redis
            image: alpine:3.11
            # Don't use alpine:latest or alpine (default tag is latest).
            imagePullPolicy: 'Always'
            ports:
              - containerPort: {{.Values.AzDProxy.cache.redisClient.targetport}}
      ```
      - Set image pull policy to always
    ```yaml
    imagePullPolicy: 'Always'
      ```