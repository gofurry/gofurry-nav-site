import { cva, type VariantProps } from 'class-variance-authority'
import { forwardRef, type ButtonHTMLAttributes } from 'react'
import { cn } from '../../lib/utils'

const buttonVariants = cva('inline-flex h-9 items-center justify-center gap-2 rounded-md px-3 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50', {
  variants: {
    variant: {
      primary: 'bg-primary text-primary-foreground hover:opacity-90',
      secondary: 'border bg-surface hover:bg-surface-muted',
      ghost: 'hover:bg-surface-muted',
      danger: 'bg-danger text-white hover:opacity-90',
    },
    size: { sm: 'h-8 px-2.5 text-xs', md: '', icon: 'size-9 px-0' },
  },
  defaultVariants: { variant: 'primary', size: 'md' },
})

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & VariantProps<typeof buttonVariants>
export const Button = forwardRef<HTMLButtonElement, ButtonProps>(({ className, variant, size, ...props }, ref) => (
  <button ref={ref} className={cn(buttonVariants({ variant, size }), className)} {...props} />
))
Button.displayName = 'Button'
