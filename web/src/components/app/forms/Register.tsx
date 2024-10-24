"use client"

import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { toast } from "@/hooks/use-toast"
import * as z from "zod"
import { Button } from "@/components/ui/button"
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, } from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { RegisterHeader } from "@/components/app/common/Header"
import { ContentLayout } from "../common/ContentLayout"
import { registerCustomer } from "@/api/auth_service"

const registerFormSchema = z.object({
  first_name: z
    .string()
    .min(2, "First name must be at least 2 characters long.")
    .max(30, "First name cannot exceed 30 characters.")
    .regex(/^[\p{L}]+$/u, "First name can only contain letters."),
  last_name: z
    .string()
    .min(2, "Last name must be at least 2 characters long.")
    .max(30, "Last name cannot exceed 30 characters.")
    .regex(/^[\p{L}]+$/u, "Last name can only contain letters.")
    .optional(),
  email: z
    .string()
    .email("Please enter a valid email address."),
  phone_number: z
    .string()
    .regex(/^\+?[1-9]\d{1,14}$/, "Please enter a valid phone number with country code."),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters long."),
  repeat_password: z
    .string()
    .min(8, "Repeat password must be at least 8 characters long."),
}).refine((data) => data.password === data.repeat_password, {
  message: "Passwords do not match. Please ensure both passwords are the same.",
  path: ["repeat_password"],
})

type AccountFormValues = z.infer<typeof registerFormSchema>

const defaultValues: AccountFormValues = {
  first_name: "",
  last_name: "",
  email: "",
  phone_number: "",
  password: "",
  repeat_password: "",
}

function Component() {
  const form = useForm<AccountFormValues>({
    resolver: zodResolver(registerFormSchema),
    defaultValues,
  })

  function onSubmit(data: AccountFormValues) {
    const dto = {
      first_name: data.first_name,
      last_name: data.last_name,
      email: data.email,
      phone_number: data.phone_number.replace("+", ""),
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

    registerCustomer(body)
  }

  return (
    <div className="flex flex-col items-center justify-center px-4 py-24 ">
      <h2 className="w-full max-w-md text-4xl font-bold m-4">Register Provider</h2>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="w-full max-w-md space-y-4 bg-white p-6 rounded-lg shadow-md border border-gray-200">
          <div className="flex space-x-4">
            <FormField
              control={form.control}
              name="first_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>First Name</FormLabel>
                  <FormControl>
                    <Input placeholder="John" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="last_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Last Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Doe" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
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
            name="phone_number"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Phone number</FormLabel>
                <FormControl>
                  <Input placeholder="+1234567890" {...field} />
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
          <FormField
            control={form.control}
            name="repeat_password"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Repeat password</FormLabel>
                <FormControl>
                  <Input type="password" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button type="submit" className="w-full font-bold">Register</Button>
        </form>
      </Form>
    </div>
  )
}

export function Register() {
  return (<>
    <RegisterHeader />
    <ContentLayout>
      <Component />
    </ContentLayout>
  </>);
}

