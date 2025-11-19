import './RequestResponseDetail.css';
import React from 'react';
import {main as models} from "../wailsjs/go/models";

interface RequestResponseDetailProps {
  data: models.RequestResponseDetail;
}

export const RequestResponseDetail: React.FC<RequestResponseDetailProps> = ({ data }) => {
  const { request, response } = data;

  const buildRequestText = (request: models.Request): string => {
    const lines: string[] = [];

    lines.push(`${request.method} ${request.url} HTTP/1.1`);
    lines.push(`Host: ${request.host}`);

    // Add headers
    request.headers.forEach((h) => {
      lines.push(`${h.name}: ${h.value}`);
    });

    // Add body if present
    if (request.body) {
      lines.push('', request.body);
    }

    return lines.join('\n');
  };

  const buildResponseText = (response: models.Response): string => {
    const lines: string[] = [];

    const statusText = '';
    lines.push(`HTTP/1.1 ${response.status_code}${statusText}`);

    // Add headers
    response.headers.forEach((h) => {
      lines.push(`${h.name}: ${h.value}`);
    });

    // Add body if present
    if (response.body) {
      lines.push('', response.body);
    }

    return lines.join('\n');
  };

  return (
    <div className="request-response-detail">
      <div className="panel request-panel">
        <div className="panel-header">Request</div>
        <pre className="panel-content">{request ? buildRequestText(request) : ""}</pre>
      </div>

      <div className="panel response-panel">
        <div className="panel-header">Response</div>
        <pre className="panel-content">{response ? buildResponseText(response) : ""}</pre>
      </div>
    </div>
  );
};
