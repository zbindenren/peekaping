import Markdown from "react-markdown";
import rehypeSanitize from "rehype-sanitize";
import { AlertTriangle, Info, AlertCircle } from "lucide-react";
import type { IncidentModel } from "@/api/types.gen";

const styleConfig: Record<
  string,
  { bg: string; border: string; icon: React.ReactNode }
> = {
  info: {
    bg: "bg-blue-50 dark:bg-blue-950/30",
    border: "border-blue-200 dark:border-blue-800",
    icon: <Info className="h-5 w-5 text-blue-500" />,
  },
  warning: {
    bg: "bg-amber-50 dark:bg-amber-950/30",
    border: "border-amber-200 dark:border-amber-800",
    icon: <AlertTriangle className="h-5 w-5 text-amber-500" />,
  },
  danger: {
    bg: "bg-red-50 dark:bg-red-950/30",
    border: "border-red-200 dark:border-red-800",
    icon: <AlertCircle className="h-5 w-5 text-red-500" />,
  },
};

const defaultStyle = styleConfig.warning;

const IncidentBanner = ({ incident }: { incident: IncidentModel }) => {
  const config = styleConfig[incident.style ?? "warning"] ?? defaultStyle;

  return (
    <div
      className={`rounded-lg border p-4 ${config.bg} ${config.border} text-left`}
    >
      <div className="flex items-start gap-3">
        <div className="mt-0.5 shrink-0">{config.icon}</div>
        <div className="min-w-0 flex-1">
          <h3 className="font-semibold">{incident.title}</h3>
          {incident.content && (
            <div className="mt-1 text-sm prose prose-sm dark:prose-invert max-w-none">
              <Markdown rehypePlugins={[rehypeSanitize]}>
                {incident.content}
              </Markdown>
            </div>
          )}
          <p className="mt-2 text-xs text-muted-foreground">
            {incident.created_at &&
              new Date(incident.created_at).toLocaleString()}
          </p>
        </div>
      </div>
    </div>
  );
};

export default IncidentBanner;
