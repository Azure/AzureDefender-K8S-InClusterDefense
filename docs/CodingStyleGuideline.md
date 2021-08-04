# Azure Defender K8S In Cluster Defense Coding Style Guideline

## Table of Contents

TBD

## Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md#specifying-map-capacity-hints)
- [Don’t just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
- [CodeReviewComments - golang/go](https://github.com/golang/go/wiki/CodeReviewComments)

## Documentation

- **Comment Sentences** -
    - See https://golang.org/doc/effective_go.html#commentary. Comments documenting declarations should be full
      sentences, even if that seems a little redundant. This approach makes them format well when extracted into godoc
      documentation.
      ```go
        // main is the entrypoint to azdproxy.
        func main() {
        webhook.StartServer()
        }
      ```
    - **When should you add it?**
        1. Packages
        2. Structs
        3. Interfaces
        4. Data members
        5. Functions
        6. Enums

## Coding style

**_- Be Consistent!_**

### Project Structure

A rule of thumb is to put each struct/interface in different go file except:

- There is only one struct that implements the interface - in this case you should write the struct and the interface in
  the same file.
- There are structs that only one struct use them - e.g. struct A probably will be in the same file with ConfigurationA

### File Structure

**File structure should be in the following order:**

1. Imports (packages)
2. Constants
3. Global variables (optional - it is best to avoid it unless you need to initialize constants on runtime)
4. Enums
5. Interfaces
6. Structs
7. Functions

**The file name should be the same as the main component in the file, e.g. the file that the Webhook struct is
implemented in, should be called 'webhook.go'**

### Packages

- Imports are organized in groups, with blank lines between them. The standard library packages are always in the first
  group - goimports tool handle this.
- **The import should be sorted** alphabetically
  ```go
    import (
        // First std lib
        "fmt"
        "os"
        // Second evreything else
        "go.uber.org/atomic"
        "golang.org/x/sync/errgroup"
    )
    ```
- **Package Names** - When naming packages, choose a name that is:
    - All lower-case. No capitals or underscores.
    - Short and succinct. Remember that the name is identified in full at every call site.
    - Not plural. For example, net/url, not net/urls.
    - Not "common", "util", "shared", or "lib". These are bad, uninformative names.

- Import aliasing - import aliases should be avoided unless there is a direct conflict between imports.

### Globals and Variables

- **Prefix Unexported Globals with _** : Prefix unexported top-level vars and constants with _ to make it clear when
  they are used that they are global symbols.
- **Group Similar Declarations** of variables.
  ```go
    const (
      a = 1
      b = 2
    )
  
  // You can group also inside a function:
  func getColor(chosenColor string) string {
      var (
        red   = color.New(0xff0000)
        green = color.New(0x00ff00)
        blue  = color.New(0x0000ff)
      )
      ...
  }
  ```

### Interfaces

- Interface should start with I and recommended ending with 'er':
  ```go
  type IWriter interface{
    Write(msg string)
  }
  ```

### Structs

- A rule of thumb is to concat the purpose of the struct as a prefix. e.g.
    ```go
        type IAnimal interface{//...}
        type Animal struct{//...}
        type DogAnimal struct{//...}
    ```

- Use Field Names to Initialize Structs:
  ```go
    type User struct{
        Name string
    } 
    
    User{Name: "Lior"}
  ```

### Functions

- Names: Mixed caps:
  ```go
  //getName is private method for get name
   func getName() (name string){}
  
   //GetName is public method for get name
   func GetName() (name string){}
   ```
- **Functions order**:
    - Functions should be sorted in rough call order
    - Order:
        1. Create/new function
        2. Receiver methods
        3. Plain utility function
            ```go 
             func newDog(name string) (Dog d)(){//...}
             func (d *dog) makeSound(){//...} 
             func dogMakeSound(d *dog){//...} 
            ```
- Named Result Parameters:
    - When there is a at least one return value from function, use () and name of the variable:
      ```go
      func getName(p *Person) (name string){//...}
        ```
- Avoid naked params -
  ```go
    printStudent("Or", true /* isGraduated */)
  ```

### Errors

- Lower case msgs: error strings should not be capitalized (unless beginning with proper nouns or acronyms) or end with
  punctuation, since they are usually printed following other context. That is, use ```fmt.Errorf("something bad")```
  not ```fmt.Errorf("Something bad")```, so that ```log.Printf("Reading %s: %v", filename, err)``` formats without a
  spurious capital letter mid-message. This does not apply to logging, which is implicitly line-oriented and not
  combined inside other messages.

### Creation

- Don't initialize object in the code, use factory/new method
  ```go
    // Bad
    config := Config{//...}
    // Good - new method
    config := config.New(//...)
    // Good - factory
    config := configFactory.Create(//...)
  ```
- When you use factory, implement interface and create method:
  ```go
    type ITracer interface{}
    type ITracerFactory interface{
        CreateTracer() *ITracer
    }
    type TracerFactory strcut{}
    type Tracer strcut{}
    func (factory *TracerFatory) CreateTracer() (tracer *Tracer){} 
  ```