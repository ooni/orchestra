import React from 'react'
import Head from 'next/head'
import Router from 'next/router'

import Autocomplete from 'react-toolbox/lib/autocomplete/Autocomplete'

import Button from 'react-toolbox/lib/button/Button'
import Input from 'react-toolbox/lib/input/Input'
import Checkbox from 'react-toolbox/lib/checkbox/Checkbox'
import Slider from 'react-toolbox/lib/slider/Slider'

import Card from 'react-toolbox/lib/card/Card'
import CardTitle from 'react-toolbox/lib/card/CardTitle'
import CardActions from 'react-toolbox/lib/card/CardActions'
import CardText from 'react-toolbox/lib/card/CardText'

import DatePicker from 'react-toolbox/lib/date_picker/DatePicker'
import TimePicker from 'react-toolbox/lib/time_picker/TimePicker'
import List from 'react-toolbox/lib/list/List'
import ListItem from 'react-toolbox/lib/list/ListItem'

import moment from 'moment'

import Layout from '../../../components/layout'
import Session from '../../../components/session'
import {
  DesignatorSlider,
  DurationPicker,
  RepeatString,
  ToScheduleString
} from '../../../components/ui/schedule'

import { Flex, Box, Grid } from 'reflexbox'


class JobCreateConfirm extends React.Component {
  static propTypes = {
    startMoment: React.PropTypes.object,
    duration: React.PropTypes.object,
    repeatCount: React.PropTypes.number,
    targetCountries: React.PropTypes.array,
    targetPlatforms: React.PropTypes.array,
    urls: React.PropTypes.string,
    alertMessage: React.PropTypes.string,
    href: React.PropTypes.string,
    altHref: React.PropTypes.string
  }

  constructor(props) {
    super(props)
  }

  render() {
    const {
      startMoment,
      duration,
      repeatCount,
      alertMessage,
      href,
      altHref,
      targetCountries,
      targetPlatforms
    } = this.props

    let durationCaption = <div>
      {RepeatString({duration, repeatCount})}
      ({ToScheduleString({
        duration: duration,
        startMoment: startMoment,
        repeatCount: repeatCount
      })})
    </div>

    let startTimeCaption = startMoment.calendar()
    startTimeCaption += ' ('
    startTimeCaption += startMoment.toString()
    startTimeCaption += ')'

    return (
      <div>
        <CardTitle title="New Alert Summary" />
        <CardText>
          <List>
          <ListItem
            ripple={false}
            caption={alertMessage}
            legend="Message" />
          <ListItem
            ripple={false}
            caption={href}
            legend="Link" />
          {altHref && altHref.split('\n')
            .map((v, i) => {
              return <ListItem
                ripple={false}
                caption={v}
                legend={`Alt Link #${i}`} />
            })
          }
          <ListItem
            ripple={false}
            caption={startTimeCaption}
            legend="Start time" />
          <ListItem
            ripple={false}
            caption={durationCaption}
            legend="Duration" />

          <ListItem
            ripple={false}
            caption={targetCountries.join(',')}
            legend="Target pountries" />
          <ListItem
            ripple={false}
            caption={targetPlatforms.join(',')}
            legend="Target platforms" />
          </List>

        </CardText>

        <style jsx>{`
        h2, h3, p, ul, div {
          margin-bottom: 16px;
        }
        `}</style>

      </div>
    )
  }
}

export default class AdminJobsAdd extends React.Component {

  constructor (props) {
    super(props)
    this.state = {
      startDate: new Date(),
      startTime: new Date(),
      startMoment: moment(),
      repeatCount: 1,
      alertMessage: '',
      href: '',
      altHref: '',
      targetCountries: [],
      targetPlatforms: [],
      duration: {W: 1},
      inputSelectorOpen: false,
      submitted: false,
      comment: '',
      session: new Session(),
      finalized: null
    }

    this.onTargetCountryChange = this.onTargetCountryChange.bind(this)
    this.onTargetPlatformChange = this.onTargetPlatformChange.bind(this)
    this.onDurationChange = this.onDurationChange.bind(this)
    this.onRepeatChange = this.onRepeatChange.bind(this)
    this.onSubmit = this.onSubmit.bind(this)
    this.onEdit = this.onEdit.bind(this)
    this.onAdd = this.onAdd.bind(this)
    this.onMessageChange = this.onMessageChange.bind(this)
    this.onHrefChange = this.onHrefChange.bind(this)
    this.onAltHrefChange = this.onAltHrefChange.bind(this)
  }

  static async getInitialProps ({req, res}) {
    // XXX get these from an API call
    const cat_codes = require('../../../static/category-codes.json')
    const countries_alpha2 = require('../../../static/countries-alpha2.json')

    let props = {}
    props.countries = {
      'any': 'All'
    }
    for (let alpha2 in countries_alpha2) {
      props.countries[alpha2] = countries_alpha2[alpha2]
    }

    props.platforms = {
      'any': 'All',
      'android': 'Android',
      'ios': 'iOS',
      'linux': 'Linux',
      'macos': 'macOS',
      'lepidopter': 'Lepidopter'
    }
    return props
  }

  componentDidMount() {
    this.state.session.redirectIfInvalid()
  }

  onMessageChange (value) {
    this.setState({ alertMessage: value})
  }
  onHrefChange (value) {
    this.setState({ href: value})
  }
  onAltHrefChange (value) {
    this.setState({ altHref: value})
  }

  onDurationChange (value) {
    this.setState({ duration: value });
  }

  onRepeatChange (value) {
    this.setState({ repeatCount: value });
  }

  onTargetCountryChange (value) {
    if (value.indexOf('any') != -1) {
      if (this.state.targetCountries.indexOf('any') != -1) {
        // If any was already there we remove it
        value.pop('any')
      } else {
        // Otherwise we clear everything else
        value = ['any']
      }
    }
    this.setState({ targetCountries: value })
  }

  onTargetPlatformChange (value) {
    if (value.indexOf('any') != -1) {
      if (this.state.targetPlatforms.indexOf('any') != -1) {
        // If any was already there we remove it
        value.pop('any')
      } else {
        // Otherwise we clear everything else
        value = ['any']
      }
    }
    this.setState({ targetPlatforms: value })
  }

  onSubmit () {
    this.setState({
      submitted: true
    })
  }

  onEdit () {
    this.setState({
      submitted: false
    })
  }

  onAdd () {
    let req = this.state.session.createRequest({baseURL: process.env.EVENTS_URL})
    let alertExtra = {}
    if (this.state.href != '') {
      alertExtra['href'] = this.state.href
      alertExtra['alt_hrefs'] = []
      if (this.state.altHref != '') {
        this.state.altHref.split("\n").forEach((v) => {
          if (v != '') {
            alertExtra['alt_hrefs'].push(v)
          }
        })
      }
    }
    let platforms = this.state.targetPlatforms.slice()
    if (platforms.indexOf('any') !== -1) {
      platforms = []
    }
    let countries = this.state.targetCountries.slice()
    if (countries.indexOf('any') !== -1) {
      countries = []
    }
    req.post('/api/v1/admin/job', {
      'schedule': ToScheduleString({
                      duration: this.state.duration,
                      startMoment: this.state.startMoment,
                      repeatCount: this.state.repeatCount
                  }),
      // XXX we currently don't set this
      'delay': 0,
      'comment': this.state.alertMessage,
      'alert': {
        'message': this.state.alertMessage,
        'extra': alertExtra,
      },
      'target': {
        'countries': countries,
        'platforms': platforms
      }
    }).then((res) => {
      this.setState({
        finalized: {
          error: null
        },
        submitted: true
      })
      Router.push('/admin/jobs')
    }).catch((err) => {
      this.setState({
        finalized: {
          error: err
        },
        submitted: true
      })
    })
  }

  render () {
    const {
      submitted,
      startDate,
      startTime,
      startMoment,
      repeatCount,
      alertMessage,
      targetCountries,
      targetPlatforms,
      duration,
      finalized
    } = this.state


    return (
      <Layout title="Add Jobs">
        <Head>
          <title>Jobs - OONI Proteus</title>
          <link href="/static/vendor/react-select.css" rel="stylesheet" />
        </Head>

        <div>
          <div className='container'>
            {submitted &&
              <Card>
              {finalized && finalized.error === null &&
                <CardTitle title="Job created successfully!" />
              }
              {finalized && finalized.error !== null &&
                <div>
                <CardTitle title="Job creation error" />
                <p>{finalized.error.toString()}</p>
                <CardActions>
                <Button
                  raised
                  onClick={this.onEdit}
                  label='Edit'/>
                <Button
                  raised
                  onClick={this.onAdd}
                  label='Retry'/>
                </CardActions>
                </div>
              }
              {!finalized && <div>
                <JobCreateConfirm
                  startMoment={this.state.startMoment}
                  duration={this.state.duration}
                  repeatCount={this.state.repeatCount}
                  duration={this.state.duration}

                  alertMessage={this.state.alertMessage}
                  href={this.state.href}
                  altHref={this.state.altHref}

                  targetCountries={this.state.targetCountries}
                  targetPlatforms={this.state.targetPlatforms}
                  urls={this.state.urls}
                  comment={this.state.comment}
                />
                <CardActions>
                <Button
                  onClick={this.onEdit}
                  label='Edit'/>
                <Button
                  onClick={this.onAdd}
                  label='Add'/>
                </CardActions>
              </div>}

              </Card>
            }
          </div>
          {!submitted &&
          <div className='scheduled-jobs container'>
            <div>
            <Card title="New Alert">

              <CardText>
                <Input
                  onChange={this.onMessageChange}
                  label="message"
                  type="text" />
                <Input
                  onChange={this.onHrefChange}
                  label="href"
                  type="text" />
                <Input
                  onChange={this.onAltHrefChange}
                  label="alt hrefs"
                  multiline
                  type="text" />
              <hr/>

              <h2>Target</h2>

              <Grid col={3} px={2}>
              <div className='option'>
                <span className='option-name'>
                  Country
                </span>
                <Autocomplete
                  direction="down"
                  selectedPosition="above"
                  label="Choose countries"
                  onChange={this.onTargetCountryChange}
                  source={this.props.countries}
                  value={this.state.targetCountries} />
              </div>
              </Grid>

              <Grid col={3} px={2}>
              <div className='option'>
                <span className='option-name'>
                  Platform
                </span>
                <Autocomplete
                  direction="down"
                  selectedPosition="above"
                  label="Choose platforms"
                  onChange={this.onTargetPlatformChange}
                  source={this.props.platforms}
                  value={this.state.targetPlatforms} />
              </div>
              </Grid>

              <hr />
              <h2>Schedule</h2>
              <Flex>
              <Box px={2}>
                <div className='option'>
                  <span className='option-name'>
                    Start on
                  </span>

                  <DatePicker
                    label="Start date"
                    autoOk={true}
                    value={this.state.startDate}
                    onChange={(startDate) => {
                      let startMoment = this.state.startMoment.clone()
                      // XXX is it correct to use UTC here?
                      startMoment.set({
                        year: startDate.getUTCFullYear(),
                        month: startDate.getUTCMonth(),
                        date: startDate.getUTCDate(),
                      })
                      this.setState({ startMoment })
                      this.setState({ startDate })
                    }}
                  />
                  <TimePicker
                    format="24hr"
                    value={this.state.startTime}
                    label="Start time"
                    onChange={(startTime) => {
                      let startMoment = this.state.startMoment.clone()
                      // XXX is it correct to use UTC here?
                      startMoment.set({
                        hour: startTime.getUTCHours(),
                        minute: startTime.getUTCMinutes(),
                        second: startTime.getUTCSeconds(),
                      })
                      this.setState({ startMoment })
                      this.setState({ startTime })
                    }}
                  />
                  <Button
                    onClick={() => {
                        this.setState({
                          startMoment: moment(new Date()),
                          startTime: new Date(),
                          startDate: new Date()
                        })
                      }
                    }
                    label='Now'/>
                </div>
              </Box>

              <Box px={2}>
                <Checkbox
                  label="Repeat"
                  checked={this.state.repeatCount !== 1}
                  onChange={(isInputChecked) => {
                    console.log(isInputChecked)
                    if (isInputChecked === true) this.onRepeatChange(2)
                    else this.onRepeatChange(1)
                  }}
                />
                {this.state.repeatCount !== 1 && <div>
                  <Checkbox
                    label="Repeat forever"
                    checked={this.state.repeatCount === 0}
                    onChange={(isInputChecked) => {
                      if (isInputChecked === true) this.onRepeatChange(0)
                      else this.onRepeatChange(2)
                    }}
                  />
                  {this.state.repeatCount !== 0 && <Input
                  type='number'
                  min='0'
                  label='times to repeat'
                  name='repeat-count'
                  value={this.state.repeatCount}
                  onChange={(value) => {this.onRepeatChange(value)}}
                  />}
                <DurationPicker onChange={this.onDurationChange}
                                duration={this.state.duration} />

                <RepeatString duration={this.state.duration} repeatCount={this.state.repeatCount} />
                </div>}
              </Box>
              </Flex>

              </CardText>
              <CardActions>
                <Button
                  raised
                  onClick={this.onSubmit}
                  label='Add' style={{marginLeft: 20}} />
              </CardActions>
            </Card>

          </div>
          </div>}
          <style jsx>{`
          .container {
            max-width: 900px;
            padding-left: 20px;
            padding-right: 20px;
            margin: auto;
          }
          hr {
            margin-top: 10px;
            margin-bottom: 10px;
            color: #ccccc;
          }
          h2 {
            font-weight: 100;
            padding-bottom: 20px;
          }
          .option-name {
            display: block;
            padding-bottom: 10px;
          }
          .option {
            padding-bottom: 20px;
          }
          .url-list {
            width: 300px;
            min-height: 100px;
          }
          `}</style>
        </div>
      </Layout>
    )
  }
}
