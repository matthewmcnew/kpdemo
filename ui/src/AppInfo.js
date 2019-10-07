import React, {useState} from "react";
import {CopyToClipboard} from 'react-copy-to-clipboard';
import {Button, Modal} from 'react-bootstrap';

export default function AppInfo(props) {
    const [show, setShow] = useState(false);

    const handleClose = () => setShow(false);
    const handleShow = () => setShow(true);

    let buildpacks = props.buildMetadata;
    if (props.buildMetadata == null) {
        buildpacks = [];
    }

    let header
    if (props.status === "True") {
        header = <h5 className={"modal-link"} onClick={handleShow}>{props.name}</h5>
    } else {
        header = <h5>{props.name}</h5>
    }

    return (
        <>

            {header}
            <div className={"team-name"}>
                Team:{props.namespace}
            </div>

            <Modal show={show} onHide={handleClose}>
                <Modal.Header closeButton>
                    <Modal.Title>Image: {props.namespace}/{props.name}</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <dt>Buildpacks</dt>
                    <dd>
                        {buildpacks.map((item, i) =>
                            <div key={i}>{item.key}:{item.version}<br/>
                            </div>
                        )}
                    </dd>
                    <br/>
                    <dt>Run Image <i className="icon-copy"/></dt>
                    <dd>
                        <div className="input-group">
                            <input type="text" className="form-control"
                                   placeholder={props.runImage} readOnly={true}/>

                            <CopyToClipboard text={props.runImage}>

                                <button className="btn btn-default" type="button" id="copy-button"
                                        title="Copy to Clipboard">
                                    Copy
                                </button>
                            </CopyToClipboard>
                        </div>
                    </dd>


                    <br/>
                    <dt>Latest Image <i className="icon-copy"/></dt>
                    <dd>
                        <div className="input-group">
                            <input type="text" className="form-control"
                                   placeholder={props.latestImage} readOnly={true}/>

                            <CopyToClipboard text={props.latestImage}>

                                <button className="btn btn-default" type="button" id="copy-button"
                                        title="Copy to Clipboard">
                                    Copy
                                </button>
                            </CopyToClipboard>
                        </div>
                    </dd>
                    <br/>


                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={handleClose}>
                        Close
                    </Button>
                </Modal.Footer>
            </Modal>
        </>
    );
}
