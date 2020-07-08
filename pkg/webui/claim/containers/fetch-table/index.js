// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React, { Component } from 'react'
import { connect } from 'react-redux'
import { push } from 'connected-react-router'
import bind from 'autobind-decorator'
import classnames from 'classnames'

import Tabular from '@ttn-lw/components/table'
import Tabs from '@ttn-lw/components/tabs'

import PropTypes from '@ttn-lw/lib/prop-types'
import debounce from '@ttn-lw/lib/debounce'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './fetch-table.styl'

const DEFAULT_PAGE = 1
const DEFAULT_TAB = 'all'
const ALLOWED_TABS = ['all']
const ALLOWED_ORDERS = ['asc', 'desc', undefined]

const filterValidator = function(filters) {
  if (!ALLOWED_TABS.includes(filters.tab)) {
    filters.tab = DEFAULT_TAB
  }

  if (!ALLOWED_ORDERS.includes(filters.order)) {
    filters.order = undefined
    filters.orderBy = undefined
  }

  if (
    (Boolean(filters.order) && !Boolean(filters.orderBy)) ||
    (!Boolean(filters.order) && Boolean(filters.orderBy))
  ) {
    filters.order = undefined
    filters.orderBy = undefined
  }

  if (!Boolean(filters.page) || filters.page < 0) {
    filters.page = DEFAULT_PAGE
  }

  return filters
}

@connect(function(state, props) {
  const base = props.baseDataSelector(state, props)

  return {
    items: base[props.entity] || [],
    totalCount: base.totalCount || 0,
    fetching: base.fetching,
    fetchingSearch: base.fetchingSearch,
    pathname: state.router.location.pathname,
  }
})
@bind
class FetchTable extends Component {
  constructor(props) {
    super(props)

    this.state = {
      query: '',
      page: 1,
      tab: 'all',
      order: undefined,
      orderBy: undefined,
    }

    const { debouncedFunction, cancel } = debounce(this.requestSearch, 350)

    this.debouncedRequestSearch = debouncedFunction
    this.debounceCancel = cancel
  }

  componentDidMount() {
    this.fetchItems()
  }

  componentWillUnmount() {
    this.debounceCancel()
  }

  fetchItems() {
    const { dispatch, pageSize, searchItemsAction, getItemsAction } = this.props

    const filters = { ...this.state, limit: pageSize }

    if (filters.query) {
      dispatch(searchItemsAction(filters))
    } else {
      dispatch(getItemsAction(filters))
    }
  }

  async onPageChange(page) {
    await this.setState(
      this.props.filterValidator({
        ...this.state,
        page,
      }),
    )

    this.fetchItems()
  }

  async requestSearch() {
    await this.setState(
      this.props.filterValidator({
        ...this.state,
        page: 1,
      }),
    )

    this.fetchItems()
  }

  async onQueryChange(query) {
    await this.setState(
      this.props.filterValidator({
        ...this.state,
        query,
      }),
    )

    this.debouncedRequestSearch()
  }

  async onOrderChange(order, orderBy) {
    await this.setState(
      this.props.filterValidator({
        ...this.state,
        order,
        orderBy,
      }),
    )

    this.fetchItems()
  }

  async onTabChange(tab) {
    await this.setState(
      this.props.filterValidator({
        ...this.state,
        tab,
      }),
    )
    this.fetchItems()
  }

  onItemClick(index) {
    const { dispatch, pathname, items, handlesPagination, pageSize } = this.props
    const { page } = this.state

    let itemIndex = index
    if (handlesPagination) {
      const pageNr = page - 1 // Switch to 0-based pagination.
      itemIndex += pageSize * pageNr
    }
    const entityPath = items[itemIndex].ids.application_id

    dispatch(push(`${pathname}${entityPath}`))
  }

  render() {
    const {
      items,
      totalCount,
      fetching,
      pageSize,
      tableTitle,
      headers,
      tabs,
      handlesPagination,
    } = this.props
    const { page, tab } = this.state

    const buttonClassNames = classnames(style.filters, {
      [style.topRule]: Boolean(tabs || tableTitle),
    })

    return (
      <div className={style.page}>
        <div className={buttonClassNames}>
          <div className={style.filtersLeft}>
            {tabs && (
              <Tabs
                active={tab}
                className={style.tabs}
                tabs={tabs}
                onTabChange={this.onTabChange}
              />
            )}
            {tableTitle && (
              <div className={style.tableTitle}>
                {tableTitle} ({totalCount})
              </div>
            )}
          </div>
        </div>
        <Tabular
          paginated
          page={page}
          totalCount={totalCount}
          pageSize={pageSize}
          onRowClick={this.onItemClick}
          onPageChange={this.onPageChange}
          loading={fetching}
          headers={headers}
          data={items}
          emptyMessage={sharedMessages.noMatch}
          handlesPagination={handlesPagination}
        />
      </div>
    )
  }
}

FetchTable.propTypes = {
  dispatch: PropTypes.func,
  fetching: PropTypes.bool,
  filterValidator: PropTypes.func,
  getItemsAction: PropTypes.func,
  handlesPagination: PropTypes.bool,
  headers: PropTypes.arrayOf(PropTypes.object),
  items: PropTypes.arrayOf(PropTypes.object),
  pageSize: PropTypes.number,
  pathname: PropTypes.string,
  searchItemsAction: PropTypes.func,
  tableTitle: PropTypes.shape({}),
  tabs: PropTypes.string,
  totalCount: PropTypes.number,
}

FetchTable.defaultProps = {
  pageSize: 20,
  handlesPagination: false,
  filterValidator,
  dispatch: () => null,
  fetching: false,
  getItemsAction: () => null,
  headers: [],
  items: [],
  pathname: '',
  searchItemsAction: () => null,
  tableTitle: {},
  tabs: '',
  totalCount: 0,
}

export default FetchTable
