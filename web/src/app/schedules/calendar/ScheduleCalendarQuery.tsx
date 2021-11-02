import React, { useEffect } from 'react'
import { gql, useQuery } from '@apollo/client'
import ScheduleCalendar from './ScheduleCalendar'
import { getStartOfWeek, getEndOfWeek } from '../../util/luxon-helpers'
import { DateTime } from 'luxon'
import { Query } from '../../../schema'
import { GenericError, ObjectNotFound } from '../../error-pages'
import { useCalendarNavigation } from './hooks'

const query = gql`
  query scheduleCalendarShifts(
    $id: ID!
    $start: ISOTimestamp!
    $end: ISOTimestamp!
    $after: String
  ) {
    userOverrides(
      input: { scheduleID: $id, start: $start, end: $end, after: $after }
    ) {
      nodes {
        id
        start
        end
        addUser {
          id
          name
        }
        removeUser {
          id
          name
        }
      }

      pageInfo {
        hasNextPage
        endCursor
      }
    }

    schedule(id: $id) {
      id
      shifts(start: $start, end: $end) {
        userID
        user {
          id
          name
        }
        start
        end
        truncated
      }

      temporarySchedules {
        start
        end
        shifts {
          userID
          user {
            id
            name
          }
          start
          end
          truncated
        }
      }
    }
  }
`

interface ScheduleCalendarQueryProps {
  scheduleID: string
}

function ScheduleCalendarQuery({
  scheduleID,
}: ScheduleCalendarQueryProps): JSX.Element | null {
  const { weekly, start } = useCalendarNavigation()

  const [queryStart, queryEnd] = weekly
    ? [
        getStartOfWeek(DateTime.fromISO(start)).toISO(),
        getEndOfWeek(DateTime.fromISO(start)).toISO(),
      ]
    : [
        getStartOfWeek(DateTime.fromISO(start).startOf('month')).toISO(),
        getEndOfWeek(DateTime.fromISO(start).endOf('month')).toISO(),
      ]

  const variables = {
    id: scheduleID,
    start: queryStart,
    end: queryEnd,
  }

  const { data, error, loading, fetchMore } = useQuery<Query>(query, {
    variables,
  })

  const hasNextPage = data?.userOverrides?.pageInfo?.hasNextPage
  const endCursor = data?.userOverrides?.pageInfo?.endCursor

  useEffect(() => {
    if (hasNextPage && endCursor) {
      console.log('fetching more')
      fetchMore({
        variables: {
          ...variables,
          after: endCursor,
        },
      }).catch((err) => console.log(err))
    }
  }, [hasNextPage, endCursor])

  console.log('num overrides: ', data?.userOverrides?.nodes.length)

  if (error) return <GenericError error={error.message} />
  if (!loading && !data?.schedule?.id) return <ObjectNotFound type='schedule' />

  return (
    <ScheduleCalendar
      scheduleID={scheduleID}
      loading={loading && !data}
      shifts={data?.schedule?.shifts ?? []}
      temporarySchedules={data?.schedule?.temporarySchedules ?? []}
      overrides={data?.userOverrides?.nodes ?? []}
    />
  )
}

export default ScheduleCalendarQuery
