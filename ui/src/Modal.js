import React from "react";
import {Button, Form, Modal} from 'react-bootstrap';

export class ConfigModal extends React.Component {

    constructor(props, context) {
        super(props, context);
        this.state = {
            show: false
        }
    }

    render() {
        const handleClose = () => this.setState({show: false});
        const handleShow = () => this.setState({show: true});

        const handleSubmit = (e) => {
            e.preventDefault();

            const buildpack = e.currentTarget.elements['buildpack'].value;
            const version = e.currentTarget.elements['version'].value;
            const runImage = e.currentTarget.elements['runImage'].value;

            this.props.setVulnerable({buildpack, version, runImage});

            handleClose();
        };

        return (
            <>
                <div className="col-lg-12">
                    <Button className={"btn-light float-right"} variant="primary" onClick={handleShow}>
                        Setup
                    </Button>
                </div>

                <Modal show={this.state.show} onHide={handleClose}>
                    <Form onSubmit={handleSubmit}>
                        <Modal.Header closeButton>
                            <Modal.Title>Mark Dependency as Vulnerable</Modal.Title>
                        </Modal.Header>

                        <Modal.Body>
                            <Form.Group>
                                <Form.Label>Run Image</Form.Label>
                                <Form.Control name="runImage" type="text" placeholder="Enter Run Image"
                                              defaultValue={this.props.vulnerable.runImage}/>
                            </Form.Group>
                        </Modal.Body>

                        <hr/>
                        <Modal.Body>
                            <Form.Group>
                                <Form.Label>Buildpack ID</Form.Label>
                                <Form.Control name="buildpack" type="text" placeholder="Enter Buildpack"
                                              defaultValue={this.props.vulnerable.buildpack}/>
                            </Form.Group>

                            <Form.Group>
                                <Form.Label>Buildpack Version</Form.Label>
                                <Form.Control name="version" type="text" placeholder="Enter Version"
                                              defaultValue={this.props.vulnerable.version}/>
                            </Form.Group>

                        </Modal.Body>
                        <Modal.Footer>
                            <Button variant="secondary" onClick={handleClose}>
                                Close
                            </Button>
                            <Button variant="primary" type="submit">
                                Save
                            </Button>
                        </Modal.Footer>
                    </Form>
                </Modal>
            </>
        );
    }

}