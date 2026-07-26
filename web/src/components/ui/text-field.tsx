import { useId, type InputHTMLAttributes } from 'react'

type TextFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string
  hint?: string
}

export function TextField({ label, hint, className = '', id, ...props }: TextFieldProps) {
  const generatedID = useId()
  const inputID = id ?? generatedID

  return (
    <label htmlFor={inputID} className={`font-body block ${className}`}>
      <span className="text-faint mb-1.5 block text-xs font-medium">{label}</span>
      <input
        id={inputID}
        className="focus-ring bg-surface text-foreground placeholder:text-faint rounded-control hover:border-border min-h-12 w-full border border-transparent px-3.5 text-sm transition-colors"
        {...props}
      />
      {hint ? <span className="text-faint mt-1.5 block text-xs">{hint}</span> : null}
    </label>
  )
}
