```cassandraql
func main() {
	defer func() func() {
		fmt.Printf("hello fucker 2\n")
		return func() {
			fmt.Printf("hello fucker 3\n")
		}
	}()()
	fmt.Printf("hello fucker 1\n")
}
```
运行结果
```cassandraql
hello fucker 2
hello fucker 1
hello fucker 3
```

```cassandraql
func main() {
	defer func() func() {
		fmt.Printf("hello fucker 2\n")
		return func() func() {
			fmt.Printf("hello fucker 3\n")
			return func() {
				fmt.Printf("hello fucker 4\n")
			}
		}()
	}()()
	fmt.Printf("hello fucker 1\n")
}
```
result
```cassandraql
hello fucker 2
hello fucker 3
hello fucker 1
hello fucker 4
```

```cassandraql
func main() {
	defer func() func() {
		fmt.Printf("hello fucker 2\n")
		return func() func() {
			fmt.Printf("hello fucker 3\n")
			return func() func() {
				fmt.Printf("hello fucker 4\n")
				return func() {
					fmt.Printf("hello fucker 5\n")
				}
			}()
		}()
	}()()
	fmt.Printf("hello fucker 1\n")
}
```
result
```cassandraql
hello fucker 2
hello fucker 3
hello fucker 4
hello fucker 1
hello fucker 5
```