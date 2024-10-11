"use client"

import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { toast } from "@/hooks/use-toast"
import * as z from "zod"
import { Button } from "@/components/ui/button"
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { LoginHeader } from "@/components/app/common/Header"
import { ContentLayout } from "../common/ContentLayout"
import { loginUser } from "@/api/auth_service"

const loginFormSchema = z.object({
  email: z
    .string()
    .email("Please enter a valid email address."),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters long."),
})

type LoginFormValues = z.infer<typeof loginFormSchema>

const defaultValues: LoginFormValues = {
  email: "",
  password: "",
};

function LoginForm() {
  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginFormSchema),
    defaultValues,
  })

  function onSubmit(data: LoginFormValues) {
    const dto = {
      email: data.email,
      password: data.password,
    }

    const body = JSON.stringify(dto)
    console.log(body)

    toast({
      title: "You submitted the following values:",
      description: (
        <pre className="mt-2 w-[340px] rounded-md bg-slate-950 p-4">
          <code className="text-white">{body}</code>
        </pre>
      ),
    })

    loginUser(body)
  }

  return (
    <div className="flex flex-col items-center justify-center px-4 py-24 ">
      <h2 className="w-full max-w-md text-4xl font-bold m-4">Login</h2>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="w-full max-w-md space-y-4 bg-white p-6 rounded-lg shadow-md border border-gray-200">
          <FormField
            control={form.control}
            name="email"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Email</FormLabel>
                <FormControl>
                  <Input type="email" placeholder="john.doe@example.com" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="password"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Password</FormLabel>
                <FormControl>
                  <Input type="password" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button type="submit" className="w-full font-bold">Login</Button>
        </form>
      </Form>
    </div>
  )
}

export function Login() {
  return (<>
    <LoginHeader />
    <ContentLayout>
      <LoginForm />
    </ContentLayout>
  </>);
}
