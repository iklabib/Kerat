using Xunit.Runners;
using System.Reflection;
using System.Text.Json;
using System.Collections.Concurrent;
using System.Security.Cryptography;
using System.Text.Json.Serialization;

public class Program
{
    public static void Main()
    {
        string assemblyPath = RandomNumberGenerator.GetHexString(8, true) + ".dll";

        string dir = AppContext.BaseDirectory;
        string target = Path.Join(dir, assemblyPath);

        // a bit hacky?
        AppDomain.CurrentDomain.AssemblyResolve += (sender, args) =>
        {
            AssemblyName assemblyName = new AssemblyName(args.Name);
            if (string.IsNullOrEmpty(assemblyName.Name))
            {
                throw new Exception("assembly name is empty");
            }

            string assemblyPath = Path.Combine(dir, assemblyName.Name + ".dll");
            if (File.Exists(assemblyPath))
            {
                return Assembly.LoadFrom(assemblyPath);
            }

            return Assembly.Load(args.Name);
        };

        var stack = new ConcurrentStack<TestResult>();

        string exec = typeof(Program).Assembly?.Location ?? "";
        using var completionEvent = new ManualResetEventSlim(false);
        using var runner = AssemblyRunner.WithoutAppDomain(exec);
        runner.OnTestFailed = info =>
        {
            stack.Push(new TestResult
            {
                Passed = false,
                Name = info.TestDisplayName,
                StackTrace = string.IsNullOrEmpty(info.ExceptionMessage) ? info.ExceptionStackTrace : info.ExceptionMessage,
            });
        };

        runner.OnTestPassed = info =>
        {
            stack.Push(new TestResult
            {
                Passed = true,
                Name = info.TestDisplayName,
            });
        };

        runner.OnExecutionComplete = _ =>
        {
            completionEvent.Set();
        };

        runner.Start();

        completionEvent.Wait();

        var testResult = stack.ToArray();

        var res = new ContainerResult 
        {
            Success = testResult.All(el => el.Passed),
            Output = testResult,
        };

        Console.Clear();
        Console.WriteLine(JsonSerializer.Serialize(res));
    }
}

public class ContainerResult 
{
    [JsonPropertyName("success")]
    public bool Success { get; set; } = false;

    [JsonPropertyName("message")]
    public string Message { get; set; } = "";

    [JsonPropertyName("output")]
    public IEnumerable<TestResult> Output { get; set; } = [];
}

public class TestResult
{
    [JsonPropertyName("passed")]
    public bool Passed { get; set; } = false;

    [JsonPropertyName("name")]
    public string Name { get; set; } = "";

    [JsonPropertyName("message")]
    public string Message { get; set; } = "";

    [JsonPropertyName("stack_trace")]
    public string StackTrace { get; set; } = "";
}
