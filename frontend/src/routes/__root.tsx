import { Link, Outlet, createRootRoute } from '@tanstack/react-router'

export const Route = createRootRoute({
  component: RootLayout,
})

// Hoisted so they aren't reallocated on every render (react-perf).
const navActiveProps = { className: 'nav-link active' }
const exactActiveOptions = { exact: true }

function RootLayout() {
  return (
    <>
      <nav className="nav">
        <Link
          to="/"
          className="nav-link"
          activeOptions={exactActiveOptions}
          activeProps={navActiveProps}
        >
          Search
        </Link>
        <Link to="/shelves" className="nav-link" activeProps={navActiveProps}>
          Shelves
        </Link>
      </nav>
      <Outlet />
    </>
  )
}
