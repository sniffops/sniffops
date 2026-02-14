import {
  LayoutDashboard,
  ListFilter,
} from 'lucide-react'
import { type SidebarData } from './types'

export const sidebarData: SidebarData = {
  navGroups: [
    {
      title: 'General',
      items: [
        {
          title: 'Dashboard',
          url: '/',
          icon: LayoutDashboard,
        },
        {
          title: 'Traces',
          url: '/traces',
          icon: ListFilter,
        },
      ],
    },
  ],
}
