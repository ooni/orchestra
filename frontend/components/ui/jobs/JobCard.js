import React from 'react'
import PropTypes from 'prop-types'

import DeleteIcon from '@material-ui/icons/Delete'
import MessageIcon from '@material-ui/icons/Message'
import AssignmentIcon from '@material-ui/icons/Assignment'

import KeyboardArrowUpIcon from '@material-ui/icons/KeyboardArrowUp'
import KeyboardArrowDownIcon from '@material-ui/icons/KeyboardArrowDown'

import Card, { CardHeader, CardContent, CardActions } from 'material-ui/Card'

import List, { ListItem, ListItemText } from 'material-ui/List'

class JobCard extends React.Component {
  static propTypes = {
    onDelete: PropTypes.func,
    comment: PropTypes.string,
    creationTime: PropTypes.string,
    delay: PropTypes.number,
    id: PropTypes.string,
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
      id,
      schedule,
      target,
      task,
      alertData,
      onDelete
    } = this.props
    const {
      isOpen
    } = this.state

    let targetCountries = 'ANY',
        targetPlatforms = 'ANY'
    if (target.countries.length > 0) {
      targetCountries = target.countries.join(',')
    }
    if (target.platforms.length > 0) {
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
    if (task) {
      cardAvatar = <Avatar><AssignmentIcon /></Avatar>
    } else {
      cardAvatar = <Avatar><MessageIcon /></Avatar>
    }
    return (
      <Card style={{position: 'relative'}}>
        <div style={{position: 'absolute', right: 0}} onClick={() => {this.setState({isOpen: !this.state.isOpen})}}>
          {isOpen && <Button><KeyboardArrowUpIcon/></Button>}
          {!isOpen && <Button><KeyboardArrowDownIcon/></Button>}
        </div>
        <CardHeader
          title={comment}
          avatar={cardAvatar}
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
          {state !== 'deleted' && <Button onClick={() => {onDelete(id)}}><DeleteIcon/></Button>}
           <Button onClick={() => {alert('I do nothing')}}>Edit</Button>
        </CardActions>
      </Card>
    )
  }
}

export default JobCard
