import { Button } from '@/components/ui/button'
import { Link } from "react-router-dom";

export function SignIn() {
  return (
    <Link to="/login">
      <Button variant="outline">Sign In</Button>
    </Link>
  )
}