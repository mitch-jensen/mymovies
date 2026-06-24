import { type Bookcase, type Shelf } from '../client'
import {
  listBookcaseShelvesOptions,
  listBookcasesOptions,
  listShelfPlacementsOptions,
} from '../client/@tanstack/react-query.gen'
import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'

export const Route = createFileRoute('/shelves')({
  component: Shelves,
})

function Shelves() {
  const { data, isFetching } = useQuery(listBookcasesOptions())
  const bookcases = data ?? []

  return (
    <main className="container">
      <h1>My Shelves</h1>

      {isFetching && bookcases.length === 0 && <p className="muted">Loading…</p>}
      {!isFetching && bookcases.length === 0 && (
        <p className="muted">No bookcases yet — add one to get started.</p>
      )}

      {bookcases.map((bookcase) => (
        <BookcaseSection key={bookcase.id} bookcase={bookcase} />
      ))}
    </main>
  )
}

function BookcaseSection({ bookcase }: { bookcase: Bookcase }) {
  const { data } = useQuery(listBookcaseShelvesOptions({ path: { bookcaseId: bookcase.id } }))
  const shelves = data ?? []

  return (
    <section className="bookcase">
      <h2 className="bookcase-name">{bookcase.name}</h2>

      {shelves.length === 0 ? (
        <p className="muted">No shelves yet.</p>
      ) : (
        shelves.map((shelf) => <ShelfRow key={shelf.id} shelf={shelf} />)
      )}
    </section>
  )
}

function ShelfRow({ shelf }: { shelf: Shelf }) {
  const { data } = useQuery(listShelfPlacementsOptions({ path: { id: shelf.id } }))
  const placements = data ?? []

  return (
    <div className="shelf">
      {placements.length === 0 ? (
        <span className="shelf-empty muted">empty</span>
      ) : (
        placements.map((item) => (
          <span key={item.release.id} className="spine" title={item.movie.title}>
            {item.movie.title}
          </span>
        ))
      )}
    </div>
  )
}
