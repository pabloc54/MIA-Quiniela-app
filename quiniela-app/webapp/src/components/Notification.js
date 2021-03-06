import { useContext } from 'react'
import { Alert } from 'react-bootstrap'
import NotificationContext from 'context/NotificationContext'

const Notification = () => {
  const { title, message, variant, visible, setVisible } = useContext(NotificationContext)

  const handleChange = () => setVisible(true)
  const handleClose = () => setVisible(false)

  return visible ? (
    <Alert variant={variant} onClose={handleClose} dismissible>
      {title && <Alert.Heading>{title}</Alert.Heading>}
      {message && <p onChange={handleChange}>{message}</p>}
    </Alert>
  ) : (
    <></>
  )
}

export default Notification
