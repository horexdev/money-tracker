import { Component } from 'react'
import type { ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(): State {
    return { hasError: true }
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex flex-col items-center justify-center h-screen gap-4 px-6 text-center">
          <p className="text-sm font-bold text-text">Something went wrong</p>
          <p className="text-xs text-muted">Please close and reopen the app.</p>
        </div>
      )
    }
    return this.props.children
  }
}
