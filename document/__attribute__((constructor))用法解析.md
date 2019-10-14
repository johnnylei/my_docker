[__attribute__((constructor))用法解析](https://www.jianshu.com/p/dd425b9dc9db)

[更多详情](https://gcc.gnu.org/onlinedocs/gcc-6.2.0/gcc/Common-Function-Attributes.html)
## attribute
GNU C 的一大特色就是__attribute__ 机制。__attribute__ 可以设置函数属性（Function Attribute ）、变量属性（Variable Attribute ）和类型属性（Type Attribute ）。

__attribute__ 书写特征是：__attribute__ 前后都有两个下划线，并切后面会紧跟一对原括弧，括弧里面是相应的__attribute__ 参数。

__attribute__ 语法格式为：__attribute__ ((attribute-list))

## attribute((constructor))
```cassandraql
int main(int argc, char * argv[]) {
    @autoreleasepool {
        printf("main function");
        return UIApplicationMain(argc, argv, nil, NSStringFromClass([AppDelegate class]));
    }
}
__attribute__((constructor)) static void beforeFunction()
{
    printf("beforeFunction\n");
}
```
运行结果
```cassandraql
beforeFunction
main function
```
所以这个__attribute__((constructor))应该是在main函数之前,执行一个函数,便于我们做一些准备工作.

『The constructor attribute causes the function to be called automatically before execution enters main (). Similarly, the destructor attribute causes the function to be called automatically after main () completes or exit () is called. Functions with these attributes are useful for initializing data that is used implicitly during the execution of the program.
』

『constructor参数让系统执行main()函数之前调用函数(被__attribute__((constructor))修饰的函数).同理, destructor让系统在main()函数退出或者调用了exit()之后,调用我们的函数.带有这些修饰属性的函数,对于我们初始化一些在程序中使用的数据非常有用.』

## 带有优先级的参数
按照文档中所说,我们还可以给属性设置优先级.这些函数并不非要写到main.m文件中,无论写到哪里,结果都是一样的.但是,为了更显式的让阅读者看到这些定义,至少,还是在main.m文件中留个声明.

```
static  __attribute__((constructor(101))) void before1()
{
    
    printf("before1\n");
}
static  __attribute__((constructor(102))) void before2()
{
    
    printf("before2\n");
}
static  __attribute__((constructor(103))) void before3()
{
    
    printf("before3\n");
}
```
上面的代码没有什么疑问.以上三个函数会依照优先级的顺序调用.另外,我以前看过,这个1-100的范围是保留的,所以,最好从100之后开始用.(但是实际上,我在项目中测试100以内的,也没有得到警告)