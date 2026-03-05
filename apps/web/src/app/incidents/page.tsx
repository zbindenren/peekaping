import { useState, useCallback, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  getIncidentsInfiniteOptions,
  getIncidentsQueryKey,
  getStatusPagesOptions,
  deleteIncidentsByIdMutation,
  patchIncidentsByIdResolveMutation,
} from "@/api/@tanstack/react-query.gen";
import IncidentCard from "./components/incident-card";
import { toast } from "sonner";
import { useNavigate } from "react-router-dom";
import type { IncidentModel } from "@/api/types.gen";
import Layout from "@/layout";
import { Label } from "@/components/ui/label";
import { useDebounce } from "@/hooks/useDebounce";
import { useSearchParams } from "@/hooks/useSearchParams";
import { Skeleton } from "@/components/ui/skeleton";
import { useIntersectionObserver } from "@/hooks/useIntersectionObserver";
import { commonMutationErrorHandler } from "@/lib/utils";
import EmptyList from "@/components/empty-list";
import { useLocalizedTranslation } from "@/hooks/useTranslation";

export default function IncidentsPage() {
  const { t } = useLocalizedTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { getParam, updateSearchParams, clearAllParams, hasParams } =
    useSearchParams();

  const [search, setSearch] = useState(getParam("q") || "");
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [showResolveDialog, setShowResolveDialog] = useState(false);
  const [pendingResolveId, setPendingResolveId] = useState<string | null>(null);
  const debouncedSearch = useDebounce(search, 300);

  useEffect(() => {
    updateSearchParams({ q: debouncedSearch });
  }, [debouncedSearch, updateSearchParams]);

  const clearAllFilters = () => {
    setSearch("");
    clearAllParams();
  };

  const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useInfiniteQuery({
      ...getIncidentsInfiniteOptions({
        query: {
          limit: 20,
          q: debouncedSearch || undefined,
        },
      }),
      getNextPageParam: (lastPage, pages) => {
        const lastLength = lastPage.data?.length || 0;
        if (lastLength < 20) return undefined;
        return pages.length;
      },
      initialPageParam: 0,
    });

  const incidents = data?.pages.flatMap((page) => page.data || []) || [];

  // Fetch status pages for displaying names on cards
  const { data: statusPagesData } = useQuery({
    ...getStatusPagesOptions(),
  });

  const statusPagesMap = new Map(
    (statusPagesData?.data || []).map((sp) => [sp.id, sp])
  );

  const deleteMutation = useMutation({
    ...deleteIncidentsByIdMutation(),
    onSuccess: () => {
      toast.success(t("incidents.deleted_success"));
      setDeleteId(null);
      queryClient.invalidateQueries({
        queryKey: getIncidentsQueryKey(),
      });
    },
    onError: commonMutationErrorHandler(t("common.error")),
  });

  const resolveMutation = useMutation({
    ...patchIncidentsByIdResolveMutation(),
    onSuccess: () => {
      toast.success(t("incidents.resolved_success"));
      setShowResolveDialog(false);
      setPendingResolveId(null);
      queryClient.invalidateQueries({
        queryKey: getIncidentsQueryKey(),
      });
    },
    onError: commonMutationErrorHandler(t("common.error")),
  });

  const handleDeleteClick = (id: string) => {
    setDeleteId(id);
  };

  const handleConfirmDelete = () => {
    if (!deleteId) return;
    deleteMutation.mutate({ path: { id: deleteId } });
  };

  const handleCancelDelete = () => {
    setDeleteId(null);
  };

  const handleResolveClick = (incident: IncidentModel) => {
    if (!incident.id) return;
    setPendingResolveId(incident.id);
    setShowResolveDialog(true);
  };

  const handleConfirmResolve = () => {
    if (!pendingResolveId) return;
    resolveMutation.mutate({ path: { id: pendingResolveId } });
  };

  const handleCancelResolve = () => {
    setShowResolveDialog(false);
    setPendingResolveId(null);
  };

  const handleCreateClick = () => {
    navigate("/incidents/new");
  };

  const handleEditClick = (id: string) => {
    navigate(`/incidents/${id}/edit`);
  };

  const handleObserver = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      const [entry] = entries;
      if (entry.isIntersecting && hasNextPage && !isFetchingNextPage) {
        fetchNextPage();
      }
    },
    [fetchNextPage, hasNextPage, isFetchingNextPage]
  );

  const { ref: sentinelRef } =
    useIntersectionObserver<HTMLDivElement>(handleObserver);

  return (
    <Layout pageName={t("incidents.title")} onCreate={handleCreateClick}>
      <div>
        <div className="mb-4 space-y-4">
          <div className="flex flex-col gap-4 sm:flex-row sm:justify-end sm:gap-4 items-end">
            {hasParams() && (
              <div className="flex justify-start">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={clearAllFilters}
                  className="w-fit h-[36px]"
                >
                  {t("common.clear_all_filters")}
                </Button>
              </div>
            )}
            <div className="flex flex-col gap-1 w-full sm:w-auto">
              <Label htmlFor="search-incidents">{t("common.search")}</Label>
              <Input
                id="search-incidents"
                placeholder={t("common.search")}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full sm:w-[400px]"
              />
            </div>
          </div>
        </div>

        {isLoading ? (
          <div className="flex flex-col space-y-2 mb-2">
            {Array.from({ length: 7 }, (_, id) => (
              <Skeleton className="h-[68px] w-full rounded-xl" key={id} />
            ))}
          </div>
        ) : incidents.length > 0 ? (
          <div>
            {incidents.map((incident: IncidentModel) => (
              <IncidentCard
                key={incident.id}
                incident={incident}
                statusPage={statusPagesMap.get(incident.status_page_id || "")}
                onClick={() =>
                  incident.id && handleEditClick(incident.id)
                }
                onDelete={() =>
                  incident.id && handleDeleteClick(incident.id)
                }
                onResolve={() => handleResolveClick(incident)}
                isPending={resolveMutation.isPending}
              />
            ))}
            <div ref={sentinelRef} style={{ height: 1 }} />
            {isFetchingNextPage && (
              <div className="flex flex-col space-y-2 mb-2">
                {Array.from({ length: 3 }, (_, i) => (
                  <Skeleton key={i} className="h-[68px] w-full rounded-xl" />
                ))}
              </div>
            )}
          </div>
        ) : (
          <EmptyList
            title={t("incidents.no_incidents")}
            text={t("incidents.no_incidents")}
            actionText={t("incidents.create")}
            onClick={() => navigate("/incidents/new")}
          />
        )}

        <AlertDialog
          open={!!deleteId}
          onOpenChange={(open) => {
            if (!open) handleCancelDelete();
          }}
        >
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {t("common.confirm_delete_title_short")}
              </AlertDialogTitle>
              <AlertDialogDescription>
                {t("incidents.confirm_delete")}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={(e) => e.stopPropagation()}>
                {t("common.cancel")}
              </AlertDialogCancel>
              <AlertDialogAction onClick={handleConfirmDelete}>
                {t("common.delete")}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>

        {showResolveDialog && (
          <AlertDialog
            open={showResolveDialog}
            onOpenChange={(open) => {
              if (!open) handleCancelResolve();
            }}
          >
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>
                  {t("incidents.resolve")}
                </AlertDialogTitle>
                <AlertDialogDescription>
                  {t("incidents.confirm_resolve")}
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel onClick={(e) => e.stopPropagation()}>
                  {t("common.cancel")}
                </AlertDialogCancel>
                <AlertDialogAction onClick={handleConfirmResolve}>
                  {t("common.confirm")}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        )}
      </div>
    </Layout>
  );
}
