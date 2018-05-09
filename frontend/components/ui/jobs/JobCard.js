import React from 'react'
import PropTypes from 'prop-types'

import DeleteIcon from '@material-ui/icons/Delete'

import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp'
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown'

import Card, { CardHeader, CardContent, CardActions } from 'material-ui/Card'

import List, { ListItem, ListItemText } from 'material-ui/List'
import Button from 'material-ui/Button'

import {
  Box
} from 'ooni-components'

class JobCard extends React.Component {
  static propTypes = {
    onDelete: PropTypes.func,
    comment: PropTypes.string,
    creationTime: PropTypes.string,
    delay: PropTypes.number,
    jobNumber: PropTypes.number,
    state: PropTypes.string,
    schedule: PropTypes.string,
    target: PropTypes.object,
    task: PropTypes.object
  }

  constructor (props) {
    super(props)
    this.state = {
      isOpen: false
    }
  }

  render () {
    const {
      state,
      comment,
      creationTime,
      delay,
      jobNumber,
      schedule,
      target,
      task,
      alertData,
      experimentData,
      onDelete
    } = this.props
    const {
      isOpen
    } = this.state

    let targetCountries = 'ANY',
        targetPlatforms = 'ANY'
    if (target.countries && target.countries.length > 0) {
      targetCountries = target.countries.join(',')
    }
    if (target.platforms && target.platforms.length > 0) {
      targetPlatforms = target.platforms.join(',')
    }
    let subtitle
    if (task) {
      subtitle = task.test_name
    }
    if (state === 'deleted') {
      subtitle = `[DELETED] ${subtitle}`
    }

    let cardAvatar
    return (
      <Box w={isOpen ? 1 : 1/3 } pl={3} pb={4}>
      <Card style={{position: 'relative'}}>
        <div style={{position: 'absolute', right: 0}} onClick={() => {this.setState({isOpen: !this.state.isOpen})}}>
          {isOpen && <Button><KeyboardArrowUpIcon/></Button>}
          {!isOpen && <Button><KeyboardArrowDownIcon/></Button>}
        </div>
        <CardHeader
          title={comment}
          subheader={subtitle}
          />
        <CardContent>
          {isOpen && <List>
            {alertData && <ListItem>
            <ListItemText
                primary={alertData.message}
                secondary="Message"/>
            </ListItem>
            }
            {alertData && <ListItem>
            <ListItemText
                primary={JSON.stringify(alertData.extra)}
                secondary="Alert Extra"/>
            </ListItem>
            }

            {task && <ListItem>
            <ListItemText
                primary={task.test_name}
                secondary="Test name"/>
            </ListItem>
            }
            {task && <ListItem>
            <ListItemText
                primary={JSON.stringify(task.arguments)}
                secondary="Test arguments"/>
            </ListItem>
            }

            <ListItem>
            <ListItemText
                primary={schedule}
                secondary="Schedule"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={''+delay}
                secondary="Delay"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={creationTime}
                secondary="Creation time"/>
            </ListItem>

            <ListItem>
            <ListItemText
                primary={targetCountries}
                secondary="Target countries"/>
            </ListItem>
            <ListItem>
            <ListItemText
                primary={targetPlatforms}
                secondary="Target platforms"/>
            </ListItem>
          </List>}
        </CardContent>
        <CardActions>
          {state !== 'deleted' && <Button onClick={() => {onDelete(jobNumber)}}><DeleteIcon/></Button>}
           <Button onClick={() => {alert('I do nothing')}}>Edit</Button>
        </CardActions>
      </Card>
      </Box>
    )
  }
}

export default JobCard
