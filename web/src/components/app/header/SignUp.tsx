import { Button } from '@/components/ui/button'
import { Link } from "react-router-dom";

export function SignUp() {
  return (
    <Link to="/register">
      <Button variant="outline">Sign Up</Button>
    </Link>
  )
}