// FILE: frontend/src/RequestResponseDetail.tsx
import './RequestResponseDetail.css';
import React from 'react';
import { main as models } from "../wailsjs/go/models";

interface RequestResponseDetailProps {
  data: models.RequestResponseDetail;
  /** Optional search term – if provided and non-empty, occurrences will be highlighted (case-insensitive) */
  search?: string;
}

export const RequestResponseDetail: React.FC<RequestResponseDetailProps> = ({ data, search = '' }) => {
  const { request, response } = data;

  const buildRequestText = (request: models.Request): string => {
    const lines: string[] = [];
    lines.push(`${request.method} ${request.url} HTTP/1.1`);
    lines.push(`Host: ${request.host}`);
    request.headers.forEach((h) => {
      lines.push(`${h.name}: ${h.value}`);
    });
    if (request.body) {
      lines.push('', request.body);
    }
    return lines.join('\n');
  };

  const buildResponseText = (response: models.Response): string => {
    const lines: string[] = [];
    lines.push(`HTTP/1.1 ${response.status_code}`);
    response.headers.forEach((h) => {
      lines.push(`${h.name}: ${h.value}`);
    });
    if (response.body) {
      lines.push('', response.body);
    }
    return lines.join('\n');
  };

  // Helper that splits text and wraps matches in a <mark> element
  const highlightText = (text: string) => {
    if (!search || search.trim() === '') {
      return <>{text}</>;
    }

    const regex = new RegExp(`(${search.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
    const parts = text.split(regex);

    return (
      <>
        {parts.map((part, i) =>
          regex.test(part) ? (
            <mark key={i}>{part}</mark>
          ) : (
            <React.Fragment key={i}>{part}</React.Fragment>
          )
        )}
      </>
    );
  };

  const requestText = request ? buildRequestText(request) : '';
  const responseText = response ? buildResponseText(response) : '';

  return (
    <div className="request-response-detail">
      <div className="panel request-panel">
        <div className="panel-header">Request</div>
        <pre className="panel-content">
          {highlightText(requestText)}
        </pre>
      </div>

      <div className="panel response-panel">
        <div className="panel-header">Response</div>
        <pre className="panel-content">
          {highlightText(responseText)}
        </pre>
      </div>
    </div>
  );
};
